#!/usr/bin/env python3
# -*- coding: UTF-8 -*-

import os, time, subprocess, sys, platform
# --debug
# --install
# --software
# --protocol
# --output
# --app
opt_debug = False
opt_install = False
opt_software = "1.7.3"

opt_app = "code-completion"
opt_os = None
opt_arch = None
opt_module = "github.com/zgsm-ai/{0}".format(opt_app)
opt_output = None
opt_cgo_enabled=0

def run_cmd(cmd):
    p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout = p.communicate()[0].decode('utf-8').strip()
    return stdout

# Get platform-specific environment variables for Go build
def get_go_env_vars():
    # Use user-specified values if provided, otherwise get from go env
    go_os = opt_os if opt_os is not None else run_cmd('go env GOOS')
    go_arch = opt_arch if opt_arch is not None else run_cmd('go env GOARCH')
    
    # Set environment variables based on current platform
    current_system = platform.system().lower()
    if current_system == "windows":
        flags = "set GOOS={0}&&set GOARCH={1}&&set CGO_ENABLED={2}&&".format(go_os, go_arch, opt_cgo_enabled)
    else:
        flags = "GOOS={0} GOARCH={1} CGO_ENABLED={2}".format(go_os, go_arch, opt_cgo_enabled)
    return flags

# Get last tag.
def last_tag():
    return run_cmd('git rev-parse --abbrev-ref HEAD')

# Get last git commit id.
def last_commit_id():
    return run_cmd('git log --pretty=format:"%h" -1')

# Assemble build command.
def build_cmd():
    build_flags = []
    if not opt_debug:
        build_flags.append("-s -w")
        
    build_flags.append("-X 'main.SoftwareVer={0}'".format(opt_software))
    last_git_tag = last_tag()
    if last_git_tag != "":
        build_flags.append("-X 'main.BuildTag={0}'".format(last_git_tag))

    commit_id = last_commit_id()
    if commit_id != "":
        build_flags.append("-X 'main.BuildCommitId={0}'".format(commit_id))

    # current time
    build_flags.append("-X 'main.BuildTime={0}'".format(time.strftime("%Y-%m-%d %H:%M:%S")))

    debugf = '-gcflags=all="-N -l"' if opt_debug else ""
    envf = get_go_env_vars()
    fext = ".exe" if platform.system().lower() == "windows" else ""
    outf = opt_output if opt_output else opt_app + fext
    subf = "install" if opt_install else "build"
    buildf = " ".join(build_flags)

    return '{0} go {1} {2} -ldflags "{3}" -o {4} .'.format(envf, subf, debugf, buildf, outf)

def parse_opts():
    global opt_debug
    global opt_install
    global opt_software
    global opt_app
    global opt_os
    global opt_arch
    global opt_output
    global opt_cgo_enabled
    argc = len(sys.argv)
    if argc == 1:
        return True
    i = 1
    while i < argc:
        arg = sys.argv[i]
        if arg == '-h':
            print("build.py [--debug] [--install] [--software VER] [--app APPNAME] [--os OS] [--arch ARCH] [--output OUTPUT] [--cgo_enabled 0/1]")
            print("  -d,--debug        编译调试版本")
            print("  -i,--install      把程序拷贝到安装目录")
            print("  -s,--software VER 指定软件版本,VER格式:x.x.x,如: 1.1.1210")
            print("  -a,--app APPNAME  当前构建的程序名字")
            print("  --os OS           指定目标操作系统,如: windows, linux, darwin")
            print("  --arch ARCH       指定目标架构,如: amd64, arm64, 386")
            print("  --output OUTPUT   指定输出文件路径")
            print("  --cgo_enabled     启用CGO,取值0或1,默认为0")
            return False
        elif arg == '-d' or arg == '--debug':
            opt_debug = True
        elif arg == '-i' or arg == '--install':
            opt_install = True
        elif arg == '-a' or arg == '--app':
            i += 1
            if i == argc:
                raise Exception("--app/-a missing parameter")
            opt_app = sys.argv[i]
        elif arg == '-s' or arg == '--software':
            i += 1
            if i == argc:
                raise Exception("--software/-s missing parameter")
            opt_software = sys.argv[i]
        elif arg == '--os':
            i += 1
            if i == argc:
                raise Exception("--os missing parameter")
            opt_os = sys.argv[i]
        elif arg == '--arch':
            i += 1
            if i == argc:
                raise Exception("--arch missing parameter")
            opt_arch = sys.argv[i]
        elif arg == '--output':
            i += 1
            if i == argc:
                raise Exception("--output missing parameter")
            opt_output = sys.argv[i]
        elif arg == '--cgo_enabled':
            i += 1
            if i == argc:
                raise Exception("--cgo_enabled missing parameter")
            value = sys.argv[i]
            if value not in ['0', '1']:
                raise Exception("--cgo_enabled value must be 0 or 1")
            opt_cgo_enabled = int(value)
        i += 1
    return True

# main
if not parse_opts():
    exit(0)
cmdline = build_cmd()
if subprocess.call(cmdline, shell=True) == 0:
    print("build ok: {0}".format(cmdline))
    exit(0)
else:
    print("build failed: {0}".format(cmdline))
    exit(1)
