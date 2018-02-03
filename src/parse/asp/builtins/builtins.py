# Contains declarations of built-in functions & objects


# Do not change the order of arguments to this function unless you feel like
# updating all of targets.go to match it.
def build_rule(name:str, cmd:str|dict='', test_cmd:str|dict='', srcs:list|dict=None, data:list=None, outs:list|dict=None,
               deps:list=None, exported_deps:list=None, secrets:list=None, tools:list|dict=None, labels:list=None,
               visibility:list=CONFIG.DEFAULT_VISIBILITY, hashes:list=None, binary:bool=False, test:bool=False,
               test_only:bool=CONFIG.DEFAULT_TESTONLY, building_description:str=None, needs_transitive_deps:bool=False,
               output_is_complete:bool=False, container:bool|dict=False, sandbox:bool=CONFIG.BUILD_SANDBOX,
               test_sandbox:bool=CONFIG.TEST_SANDBOX, no_test_output:bool=False, flaky:bool|int=0, build_timeout:int=0,
               test_timeout:int=0, pre_build:function=None, post_build:function=None, requires:list=None, provides:dict=None,
               licences:list=CONFIG.DEFAULT_LICENCES, test_outputs:list=None, system_srcs:list=None, stamp:bool=False,
               tag:str='', optional_outs:list=None):
    pass


def len(obj):
    pass
def enumerate(seq:list):
    pass
def zip(args):
    pass

def join(self, seq:list):
    pass


def split(self, on:str=' '):
    pass
def replace(self, old:str, new:str):
    pass
def partition(self, sep:str):
    pass
def rpartition(self, sep:str):
    pass
def startswith(self, s:str):
    pass
def endswith(self, s:str):
    pass
def format(self):
    pass
def lstrip(self, cutset:str):
    pass
def rstrip(self, cutset:str):
    pass
def strip(self, cutset:str=' '):
    return self.lstrip(cutset).rstrip(cutset)
def find(self, needle:str):
    pass
def rfind(self, needle:str):
    pass


def subinclude(target:str, hash:str=None):
    pass


def isinstance(obj, types:function|list):
    pass


def callable(obj):
    pass


def range(start:int, stop:int=None):
    pass


def bool(b):
    pass


def int(i):
    raise 'int is not callable'


def str(s):
    pass


def list(l):
    raise 'list is not callable'


def dict(d):
    raise 'dict is not callable'


def glob(includes:list, excludes:list=[], exclude:list=None, hidden:bool=False):
    pass


def package():
    pass


def sorted(seq:list):
    pass


def get(self, key:str, default=None):
    pass


def setdefault(self, key:str, default=None):
    if key in self:
        return self[key]
    self[key] = default
    return default


def get_base_path():
    return PACKAGE_NAME


def package_name():
    return PACKAGE_NAME


def keys(self):
    pass
def values(self):
    pass
def items(self):
    pass
def copy(self):
    pass


def debug(args):
    pass
def info(args):
    pass
def notice(args):
    pass
def warning(args):
    pass
def error(args):
    pass
def fatal(args):
    pass

log = {
    'debug': debug,
    'info': info,
    'notice': notice,
    'warning': warning,
    'error': error,
    'fatal': fatal,
}


def join_path(paths:str):
    pass  # Has to be implemented natively since it's varargs.


def split_path(p:str):
    before, _, after = p.rpartition('/')
    return before, after


def splitext(p:str):
    before, _, after = p.rpartition('.')
    return before, after


def basename(p:str):
    """Returns the final component of a pathname"""
    _, _, after = p.rpartition('/')
    return after


def dirname(p:str):
    """Returns the directory component of a pathname"""
    before, _, after = p.rpartition('/')
    return before or after


# Exception types.
# There is no actual Exception, but str() is similar enough here.
ParseError = str
ConfigError = str


# Post-build callback functions.
# TODO(peterebden): Can we stop these from being called at other times?
def get_labels(name:str, prefix:str):
    pass
def has_label(name:str, prefix:str):
    return len(get_labels(name, prefix)) > 0
def add_dep(target:str, dep:str, exported:bool=False):
    pass
def add_exported_dep(target:str, dep:str):
    add_dep(target, dep, True)
def add_out(target:str, name:str, out:str=''):
    pass
def add_licence(target:str, licence:str):
    pass
def get_command(target:str, config:str=''):
    pass
def set_command(target:str, config:str, command:str=''):
    pass
