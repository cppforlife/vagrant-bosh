package vm

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

const runitProvisionerLogTag = "RunitProvisioner"

var (
	// Matches 'svlogd -tt /var/vcap/bosh/log'
	runitSvlogdRegex = regexp.MustCompile(`\s*svlogd\s+\-tt\s+(.+)\s*`)

	// Matches 'down: agent: 3s, normally up; run: log: (pid 15318) 7762s'
	runitStatusDownRegex = regexp.MustCompile(`\Adown: [a-z\/]+: \d+`)
)

// RunitProvisioner installs runit via apt-get and
// adds specified service under runit's control.
type RunitProvisioner struct {
	fs           boshsys.FileSystem
	cmds         SimpleCmds
	runner       boshsys.CmdRunner
	assetManager AssetManager
	logger       boshlog.Logger
}

func NewRunitProvisioner(
	fs boshsys.FileSystem,
	cmds SimpleCmds,
	runner boshsys.CmdRunner,
	assetManager AssetManager,
	logger boshlog.Logger,
) RunitProvisioner {
	return RunitProvisioner{
		fs:           fs,
		cmds:         cmds,
		runner:       runner,
		assetManager: assetManager,
		logger:       logger,
	}
}

func (p RunitProvisioner) Provision(name string) error {
	err := p.installRunit()
	if err != nil {
		return bosherr.WrapError(err, "Installing runit")
	}

	err = p.setUpService(name)
	if err != nil {
		return bosherr.WrapError(err, "Setting up service")
	}

	return nil
}

func (p RunitProvisioner) installRunit() error {
	p.logger.Info(runitProvisionerLogTag, "Installing runit")

	// todo non-bash
	cmd := boshsys.Command{
		Name: "bash",
		Args: []string{
			"-c", "apt-get -q -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' install runit",
		},
		Env: map[string]string{
			"DEBIAN_FRONTEND": "noninteractive",
		},
	}

	_, _, _, err := p.runner.RunComplexCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (p RunitProvisioner) setUpService(name string) error {
	p.logger.Info(runitProvisionerLogTag, "Setting up %s service", name)

	servicePath := fmt.Sprintf("/etc/sv/%s", name)
	enableServicePath := fmt.Sprintf("/etc/service/%s", name)

	err := p.stopRunAndLog(servicePath, enableServicePath, name)
	if err != nil {
		return bosherr.WrapError(err, "Stopping run and log")
	}

	err = p.setUpRun(servicePath, name)
	if err != nil {
		return bosherr.WrapError(err, "Setting up run")
	}

	err = p.setUpLog(servicePath, name)
	if err != nil {
		return bosherr.WrapError(err, "Setting up log")
	}

	err = p.startRunAndLog(servicePath, enableServicePath, name)
	if err != nil {
		return bosherr.WrapError(err, "Starting run and log")
	}

	return nil
}

// setUpRun sets up script that runit will execute for the primary process
func (p RunitProvisioner) setUpRun(servicePath, name string) error {
	err := p.cmds.MkdirP(servicePath)
	if err != nil {
		return err
	}

	runPath := fmt.Sprintf("%s/run", servicePath)

	err = p.assetManager.Place(fmt.Sprintf("%s/%s-run", name, name), runPath)
	if err != nil {
		return err
	}

	return p.cmds.ChmodX(runPath)
}

// setUpLog sets up logging destination for the service
func (p RunitProvisioner) setUpLog(servicePath, name string) error {
	logPath := fmt.Sprintf("%s/log", servicePath)

	err := p.cmds.MkdirP(logPath)
	if err != nil {
		return err
	}

	logRunPath := fmt.Sprintf("%s/run", logPath)

	err = p.assetManager.Place(fmt.Sprintf("%s/%s-log", name, name), logRunPath)
	if err != nil {
		return err
	}

	err = p.cmds.ChmodX(logRunPath)
	if err != nil {
		return err
	}

	contens, err := p.fs.ReadFileString(logRunPath)
	if err != nil {
		return err
	}

	// First match is the whole string
	svlogdPaths := runitSvlogdRegex.FindStringSubmatch(contens)

	// Create log file destination so that runit process can properly log
	if len(svlogdPaths) == 2 {
		err = p.cmds.MkdirP(svlogdPaths[1])
		if err != nil {
			return err
		}
	}

	return nil
}

func (p RunitProvisioner) stopRunAndLog(servicePath, enableServicePath, name string) error {
	err := p.stopRunsv(name)
	if err != nil {
		return bosherr.WrapError(err, "Stopping service")
	}

	err = p.stopRunsv(fmt.Sprintf("%s/log", name))
	if err != nil {
		return bosherr.WrapError(err, "Stopping log service")
	}

	err = p.fs.RemoveAll(enableServicePath)
	if err != nil {
		return err
	}

	// Clear out all service state kept in supervise/ and control/ dirs
	return p.fs.RemoveAll(servicePath)
}

func (p RunitProvisioner) startRunAndLog(servicePath, enableServicePath, name string) error {
	// Enabling service will kick in monitoring
	_, _, _, err := p.runner.RunCommand("ln", "-sf", servicePath, enableServicePath)

	return err
}

func (p RunitProvisioner) stopRunsv(name string) error {
	p.logger.Info(runitProvisionerLogTag, "Stopping runsv")

	downStdout, _, _, err := p.runner.RunCommand("sv", "down", name)
	if err != nil {
		p.logger.Error(runitProvisionerLogTag, "Ignoring down error %s", err.Error())
	}

	// If runsv configuration does not exist, service was never started
	if strings.Contains(downStdout, "file does not exist") {
		return nil
	}

	var lastStatusStdout string

	for i := 0; i < 20; i++ {
		lastStatusStdout, _, _, _ = p.runner.RunCommand("sv", "status", name)

		if runitStatusDownRegex.MatchString(lastStatusStdout) {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return bosherr.New("Failed to stop runsv for %s. Output: %s", name, lastStatusStdout)
}
