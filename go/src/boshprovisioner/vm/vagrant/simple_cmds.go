package vagrant

import (
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

type SimpleCmds struct {
	runner boshsys.CmdRunner
	logger boshlog.Logger
}

func NewSimpleCmds(
	runner boshsys.CmdRunner,
	logger boshlog.Logger,
) SimpleCmds {
	return SimpleCmds{
		runner: runner,
		logger: logger,
	}
}

func (r SimpleCmds) MkdirP(path string) error {
	return r.run("mkdir", "-p", path)
}

func (r SimpleCmds) ChmodX(path string) error {
	return r.run("chmod", "+x", path)
}

func (r SimpleCmds) Touch(path string) error {
	return r.run("touch", path)
}

func (r SimpleCmds) Mv(srcPath, dstPath string) error {
	return r.run("mv", srcPath, dstPath)
}

func (r SimpleCmds) Chmod(mode, path string) error {
	return r.run("chmod", mode, path)
}

func (r SimpleCmds) Chown(user, group, path string) error {
	return r.run("chmod", user+":"+group, path)
}

func (r SimpleCmds) Bash(script string) error {
	return r.run("bash", "-c", script)
}

func (r SimpleCmds) run(cmd string, args ...string) error {
	_, _, _, err := r.runner.RunCommand(cmd, args...)
	return err
}
