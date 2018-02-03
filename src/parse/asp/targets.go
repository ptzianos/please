package asp

import (
	"os"
	"path"
	"strings"
	"time"

	"core"
)

// filegroupCommand is the command we put on filegroup rules.
const filegroupCommand = pyString("filegroup")

// hashFilegroupCommand is similarly the command for hash_filegroup rules.
const hashFilegroupCommand = pyString("hash_filegroup")

// createTarget creates a new build target as part of build_rule().
func createTarget(s *scope, args []pyObject) *core.BuildTarget {
	isTruthy := func(i int) bool {
		return args[i] != nil && args[i] != False && (args[i] == &True || args[i].IsTruthy())
	}

	name := string(args[0].(pyString))
	testCmd := args[2]
	container := isTruthy(19)
	test := isTruthy(14)

	// A bunch of error checking first
	s.NAssert(name == "all", "'all' is a reserved build target name.")
	s.NAssert(name == "", "Target name is empty")
	s.NAssert(strings.ContainsRune(name, '/'), "/ is a reserved character in build target names")
	s.NAssert(strings.ContainsRune(name, ':'), ": is a reserved character in build target names")
	s.NAssert(container && !test, "Only tests can have container=True")

	if tag := args[34]; tag != nil {
		if tagStr := string(tag.(pyString)); tagStr != "" {
			name = tagName(name, tagStr)
		}
	}
	label, err := core.TryNewBuildLabel(s.pkg.Name, name)
	s.Assert(err == nil, "Invalid build target name %s", name)

	target := core.NewBuildTarget(label)
	target.IsBinary = isTruthy(13)
	target.IsTest = test
	target.NeedsTransitiveDependencies = isTruthy(17)
	target.OutputIsComplete = isTruthy(18)
	target.Containerise = container
	target.Sandbox = isTruthy(20)
	target.TestOnly = isTruthy(15)
	if timeout := args[24]; timeout != nil {
		target.BuildTimeout = time.Duration(timeout.(pyInt)) * time.Second
	}
	target.Stamp = isTruthy(33)
	target.IsHashFilegroup = args[1] == hashFilegroupCommand
	target.IsFilegroup = args[1] == filegroupCommand || target.IsHashFilegroup
	if desc := args[16]; desc != nil && desc != None {
		target.BuildingDescription = string(desc.(pyString))
	}
	if target.IsBinary {
		target.AddLabel("bin")
	}
	target.Command, target.Commands = decodeCommands(s, args[1])
	if test {
		if flaky := args[23]; flaky != nil {
			if flaky == True {
				target.Flakiness = 3
				target.AddLabel("flaky") // Automatically label flaky tests
			} else if i, ok := flaky.(pyInt); ok {
				target.Flakiness = int(i)
				target.AddLabel("flaky")
			}
		}
		// Automatically label containerised tests.
		if target.Containerise {
			target.AddLabel("container")
		}
		if testCmd != nil && testCmd != None {
			target.TestCommand, target.TestCommands = decodeCommands(s, args[2])
		}
		if timeout := args[25]; timeout != nil {
			target.TestTimeout = time.Duration(timeout.(pyInt)) * time.Second
		}
		target.TestSandbox = isTruthy(21)
		target.NoTestOutput = isTruthy(22)
	}
	return target
}

// decodeCommands takes a Python object and returns it as a string and a map; only one will be set.
func decodeCommands(s *scope, obj pyObject) (string, map[string]string) {
	if obj == nil || obj == None {
		return "", nil
	} else if cmd, ok := obj.(pyString); ok {
		return string(cmd), nil
	}
	cmds, ok := obj.(pyDict)
	s.Assert(ok, "Unknown type for command [%s]", obj.Type())
	// Have to convert all the keys too
	m := make(map[string]string, len(cmds))
	for k, v := range cmds {
		if v != None {
			sv, ok := v.(pyString)
			s.Assert(ok, "Unknown type for command")
			m[k] = string(sv)
		}
	}
	return "", m
}

// populateTarget sets the assorted attributes on a build target.
func populateTarget(s *scope, t *core.BuildTarget, args []pyObject) {
	addMaybeNamed(s, "srcs", args[3], t.AddSource, t.AddNamedSource, false, false)
	addMaybeNamed(s, "tools", args[9], t.AddTool, t.AddNamedTool, true, true)
	addMaybeNamed(s, "system_srcs", args[32], t.AddSource, nil, true, false)
	addMaybeNamed(s, "data", args[4], t.AddDatum, nil, false, false)
	addMaybeNamedOutput(s, "outs", args[5], t.AddOutput, t.AddNamedOutput, t, false)
	addMaybeNamedOutput(s, "optional_outs", args[35], t.AddOptionalOutput, nil, t, true)
	addMaybeNamedOutput(s, "test_outputs", args[31], t.AddTestOutput, nil, t, false)
	addDependencies(s, "deps", args[6], t, false)
	addDependencies(s, "exported_deps", args[7], t, true)
	addStrings(s, "labels", args[10], t.AddLabel)
	addStrings(s, "hashes", args[12], t.AddHash)
	addStrings(s, "licences", args[30], t.AddLicence)
	addStrings(s, "requires", args[28], t.AddRequire)
	addStrings(s, "visibility", args[11], func(str string) {
		t.Visibility = append(t.Visibility, parseVisibility(s, str))
	})
	addStrings(s, "secrets", args[8], func(str string) {
		s.NAssert(strings.HasPrefix(str, "//"), "Secret %s of %s cannot be a build label", str, t.Label.Name)
		s.Assert(strings.HasPrefix(str, "/") || strings.HasPrefix(str, "~"), "Secret '%s' of %s is not an absolute path", str, t.Label.Name)
		t.Secrets = append(t.Secrets, str)
	})
	addProvides(s, "provides", args[29], t)
	setContainerSettings(s, "container", args[19], t)
	if f := callbackFunction(s, "pre_build", args[26], 1, "argument"); f != nil {
		t.NewPreBuildFunction = &preBuildFunction{f: f, s: s}
	}
	if f := callbackFunction(s, "post_build", args[27], 2, "arguments"); f != nil {
		t.NewPostBuildFunction = &postBuildFunction{f: f, s: s}
	}
}

// addMaybeNamed adds inputs to a target, possibly in named groups.
func addMaybeNamed(s *scope, name string, obj pyObject, anon func(core.BuildInput), named func(string, core.BuildInput), systemAllowed, tool bool) {
	if obj == nil {
		return
	}
	if l, ok := obj.(pyList); ok {
		for _, li := range l {
			if bi := parseBuildInput(s, li, name, systemAllowed, tool); bi != nil {
				anon(bi)
			}
		}
	} else if d, ok := obj.(pyDict); ok {
		s.Assert(named != nil, "%s cannot be given as a dict", name)
		for k, v := range d {
			if v != None {
				l, ok := v.(pyList)
				s.Assert(ok, "Values of %s must be lists of strings", name)
				for _, li := range l {
					if bi := parseBuildInput(s, li, name, systemAllowed, tool); bi != nil {
						named(k, bi)
					}
				}
			}
		}
	} else if obj != None {
		s.Assert(false, "Argument %s must be a list or dict, not %s", name, obj.Type())
	}
}

// addMaybeNamedOutput adds outputs to a target, possibly in a named group
func addMaybeNamedOutput(s *scope, name string, obj pyObject, anon func(string), named func(string, string), t *core.BuildTarget, optional bool) {
	if obj == nil {
		return
	}
	if l, ok := obj.(pyList); ok {
		for _, li := range l {
			out, ok := li.(pyString)
			s.Assert(ok, "outs must be strings")
			anon(string(out))
			if !optional || !strings.HasPrefix(string(out), "*") {
				s.pkg.MustRegisterOutput(string(out), t)
			}
		}
	} else if d, ok := obj.(pyDict); ok {
		s.Assert(named != nil, "%s cannot be given as a dict", name)
		for k, v := range d {
			l, ok := v.(pyList)
			s.Assert(ok, "Values must be lists of strings")
			for _, li := range l {
				out, ok := li.(pyString)
				s.Assert(ok, "outs must be strings")
				named(k, string(out))
				if !optional || !strings.HasPrefix(string(out), "*") {
					s.pkg.MustRegisterOutput(string(out), t)
				}
			}
		}
	} else if obj != None {
		s.Assert(false, "Argument %s must be a list or dict, not %s", name, obj.Type())
	}
}

// addDependencies adds dependencies to a target, which may or may not be exported.
func addDependencies(s *scope, name string, obj pyObject, target *core.BuildTarget, exported bool) {
	addStrings(s, name, obj, func(str string) {
		label := core.ParseBuildLabel(str, s.pkg.Name)
		target.AddMaybeExportedDependency(label, exported, false)
	})
}

// addStrings adds an arbitrary set of strings to the target (e.g. labels etc).
func addStrings(s *scope, name string, obj pyObject, f func(string)) {
	if obj != nil && obj != None {
		l, ok := obj.(pyList)
		s.Assert(ok, "Argument %s must be a list, not %s", name, obj.Type())
		for _, li := range l {
			str, ok := li.(pyString)
			s.Assert(ok, "%s must be strings", name)
			if str != "" {
				f(string(str))
			}
		}
	}
}

// addProvides adds a set of provides to the target, which is a dict of string -> label
func addProvides(s *scope, name string, obj pyObject, t *core.BuildTarget) {
	if obj != nil && obj != None {
		d, ok := obj.(pyDict)
		s.Assert(ok, "Argument %s must be a dict, not %s", name, obj.Type())
		for k, v := range d {
			str, ok := v.(pyString)
			s.Assert(ok, "%s keys must be strings", name)
			t.AddProvide(k, core.ParseBuildLabel(string(str), s.pkg.Name))
		}
	}
}

// setContainerSettings sets any custom container settings on the target.
func setContainerSettings(s *scope, name string, obj pyObject, t *core.BuildTarget) {
	if obj != nil && obj != None && obj != True && obj != False {
		d, ok := obj.(pyDict)
		s.Assert(ok, "Argument %s must be a dict, not %s", name, obj.Type())
		for k, v := range d {
			str, ok := v.(pyString)
			s.Assert(ok, "%s keys must be strings", name)
			err := t.SetContainerSetting(strings.Replace(k, "_", "", -1), string(str))
			s.Assert(err == nil, "%s", err)
		}
	}
}

// parseVisibility converts a visibility string to a build label.
// Mostly they are just build labels but other things are allowed too (e.g. "PUBLIC").
func parseVisibility(s *scope, vis string) core.BuildLabel {
	if vis == "PUBLIC" || (s.state.Config.Bazel.Compatibility && vis == "//visibility:public") {
		return core.WholeGraph[0]
	}
	return core.ParseBuildLabel(vis, s.pkg.Name)
}

func parseBuildInput(s *scope, in pyObject, name string, systemAllowed, tool bool) core.BuildInput {
	src, ok := in.(pyString)
	if !ok {
		s.Assert(in == None, "Items in %s must be strings", name)
		return nil
	}
	return parseSource(s, string(src), systemAllowed, tool)
}

// parseSource parses an incoming source label as either a file or a build label.
// Identifies if the file is owned by this package and returns an error if not.
func parseSource(s *scope, src string, systemAllowed, tool bool) core.BuildInput {
	if core.LooksLikeABuildLabel(src) {
		return core.MustParseNamedOutputLabel(src, s.pkg.Name)
	}
	s.Assert(src != "", "Empty source path")
	s.Assert(!strings.Contains(src, "../"), "%s is an invalid path; build target paths can't contain ../", src)
	if src[0] == '/' || src[0] == '~' {
		s.Assert(systemAllowed, "%s is an absolute path; that's not allowed", src)
		return core.SystemFileLabel{Path: src}
	} else if strings.Contains(src, "/") {
		// Target is in a subdirectory, check nobody else owns that.
		for dir := path.Dir(path.Join(s.pkg.Name, src)); dir != s.pkg.Name && dir != "."; dir = path.Dir(dir) {
			s.Assert(!core.IsPackage(dir), "Trying to use file %s, but that belongs to another package (%s)", src, dir)
		}
	} else if tool {
		// "go" as a source is interpreted as a file, as a tool it's interpreted as something on the PATH.
		return core.SystemPathLabel{Name: src, Path: s.state.Config.Build.Path}
	}
	// Make sure it's not the actual build file.
	for _, filename := range core.State.Config.Parse.BuildFileName {
		s.Assert(filename != src, "You can't specify the BUILD file as an input to a rule")
	}
	return core.FileLabel{File: src, Package: s.pkg.Name}
}

// callbackFunction extracts a pre- or post-build function for a target.
func callbackFunction(s *scope, name string, obj pyObject, requiredArguments int, arguments string) *pyFunc {
	if obj != nil && obj != None {
		f := obj.(*pyFunc)
		s.Assert(len(f.args) == requiredArguments, "%s callbacks must take exactly %d %s (%s takes %d)", name, requiredArguments, arguments, f.name, len(f.args))
		return f
	}
	return nil
}

// A preBuildFunction implements the core.PreBuildFunction interface
type preBuildFunction struct {
	f *pyFunc
	s *scope
}

func (f *preBuildFunction) Call(target *core.BuildTarget) error {
	s := f.f.scope.NewScope()
	s.Callback = true
	s.Set(f.f.args[0], pyString(target.Label.Name))
	return annotateCallbackError(s, target, s.interpreter.interpretStatements(s, f.f.code))
}

// A postBuildFunction implements the core.PostBuildFunction interface
type postBuildFunction struct {
	f *pyFunc
	s *scope
}

func (f *postBuildFunction) Call(target *core.BuildTarget, output string) error {
	log.Debug("Running post-build function for %s. Build output:\n%s", target.Label, output)
	s := f.f.scope.NewScope()
	s.Callback = true
	s.Set(f.f.args[0], pyString(target.Label.Name))
	s.Set(f.f.args[1], fromStringList(strings.Split(strings.TrimSpace(output), "\n")))
	return annotateCallbackError(s, target, s.interpreter.interpretStatements(s, f.f.code))
}

// annotateCallbackError adds some information to an error on failure about where it was in the file.
func annotateCallbackError(s *scope, target *core.BuildTarget, err error) error {
	if err == nil {
		return nil
	}
	// Something went wrong, find the BUILD file and attach some info.
	pkg := s.state.Graph.Package(target.Label.PackageName)
	f, _ := os.Open(pkg.Filename)
	return s.interpreter.parser.annotate(err, f)
}
