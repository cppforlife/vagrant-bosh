package vm

import (
	"fmt"
	"strings"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpeventlog "boshprovisioner/eventlog"
)

const (
	depsProvisionerLogTag = "DepsProvisioner"
	runAptGetUpdateMsg    = "E: Unable to fetch some archives, maybe run apt-get update"
)

// DepsProvisioner installs basic dependencies for running
// packaging scripts from BOSH packages. It also installs
// non-captured dependencies by few common BOSH releases.
// (e.g. cmake, quota)
type DepsProvisioner struct {
	runner   boshsys.CmdRunner
	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewDepsProvisioner(
	runner boshsys.CmdRunner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) DepsProvisioner {
	return DepsProvisioner{
		runner:   runner,
		eventLog: eventLog,
		logger:   logger,
	}
}

func (p DepsProvisioner) Provision() error {
	pkgNames := []string{
		// For packaging scripts in BOSH releases
		"build-essential", // 16sec
		"cmake",           // 6sec
		"libcap-dev",      // 3sec

		"libbz2-1.0",   // noop on precise64 Vagrant box
		"libbz2-dev",   // 2sec
		"libxslt1-dev", // 2sec
		"libxml2-dev",  // 2sec

		// For warden
		"quota", // 1sec
	}

	stage := p.eventLog.BeginStage("Installing dependencies", len(pkgNames))

	installedPkgs, err := p.listInstalledPkgs()
	if err != nil {
		return bosherr.WrapError(err, "Listing installed packages")
	}

	for _, pkgName := range pkgNames {
		task := stage.BeginTask(fmt.Sprintf("Package %s", pkgName))

		if p.isPkgInstalled(pkgName, installedPkgs) {
			p.logger.Debug(depsProvisionerLogTag, "Package %s is already installed", pkgName)
			task.End(nil)
			continue
		}

		err := task.End(p.installPkg(pkgName))
		if err != nil {
			return bosherr.WrapError(err, "Installing %s", pkgName)
		}
	}

	return nil
}

func (p DepsProvisioner) installPkg(name string) error {
	p.logger.Debug(depsProvisionerLogTag, "Installing package %s", name)

	_, _, _, err := p.runner.RunCommand("apt-get", "-y", "install", name)
	if err == nil {
		return nil
	}

	// Avoid running 'apt-get update' since it usually takes 30sec
	if strings.Contains(err.Error(), runAptGetUpdateMsg) {
		_, _, _, err := p.runner.RunCommand("apt-get", "-y", "update")
		if err != nil {
			return bosherr.WrapError(err, "Updating sources")
		}

		// Try second time after updating
		_, _, _, err = p.runner.RunCommand("apt-get", "-y", "install", name)
		if err != nil {
			return bosherr.WrapError(err, "Installing %s after updating", name)
		}
	}

	return nil
}

func (p DepsProvisioner) listInstalledPkgs() ([]string, error) {
	var installedPkgs []string

	installedPkgStdout, _, _, err := p.runner.RunCommand("dpkg", "--get-selections")
	if err != nil {
		return installedPkgs, bosherr.WrapError(err, "dkpg query")
	}

	for _, line := range strings.Split(installedPkgStdout, "\n") {
		pieces := strings.Fields(line)

		// Last line is empty
		if len(pieces) == 2 && pieces[1] == "install" {
			installedPkgs = append(installedPkgs, pieces[0])
		}
	}

	return installedPkgs, nil
}

func (p DepsProvisioner) isPkgInstalled(pkgName string, installedPkgs []string) bool {
	for _, installedPkgName := range installedPkgs {
		if pkgName == installedPkgName {
			return true
		}
	}

	return false
}
