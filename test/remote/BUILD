# Various helpers for testing remote execution.
remote_file(
    name = "elan",
    binary = True,
    hashes = ["9f67d54ebf96e7e552db6d5c10811ed814a89191ae425a2d658ea7ec9bbe9700"],
    url = "https://github.com/thought-machine/please-servers/releases/download/v6.0.3/elan",
)

remote_file(
    name = "mettle",
    binary = True,
    hashes = ["e723d841b1a5b60ffa32f6bacf2161eb69e17827ea526c9965948d8363b86123"],
    url = "https://github.com/thought-machine/please-servers/releases/download/v6.0.3/mettle",
)

remote_file(
    name = "zeal",
    binary = True,
    hashes = ["835559bcacfc2365f3ddb082ac027d96290843413c2f09f20f51499918ae35c6"],
    url = "https://github.com/thought-machine/please-servers/releases/download/v6.0.3/zeal",
)

sh_cmd(
    name = "run_elan",
    srcs = [":elan"],
    cmd = "mkdir -p plz-out/elan && exec $(out_location :elan) -s file://\\\\$PWD/plz-out/elan -v warning --log_file plz-out/log/elan.log --port 7777 --admin_disabled",
)

sh_cmd(
    name = "run_mettle",
    srcs = [":mettle"],
    cmd = "exec $(out_location :mettle) dual -s 127.0.0.1:7777 -d plz-out/mettle -v warning --log_file plz-out/log/mettle.log --browser http://127.0.0.1:7779 --num_workers 8 --port 7778 --admin_disabled",
)

sh_cmd(
    name = "run_zeal",
    srcs = [":zeal"],
    cmd = "exec $(out_location :zeal) -v warning -s 127.0.0.1:7777 --log_file plz-out/log/zeal.log --port 7776 --admin_disabled",
)
