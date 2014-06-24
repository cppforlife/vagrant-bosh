package vagrant

import (
	"fmt"
	"strings"
	"time"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpeventlog "boshprovisioner/eventlog"
)

const (
	depsProvisionerLogTag          = "DepsProvisioner"
	depsProvisionerAptGetUpdateMsg = "E: Unable to fetch some archives, maybe run apt-get update"
)

// DepsProvisioner installs basic dependencies for running
// packaging scripts from BOSH packages. It also installs
// non-captured dependencies by few common BOSH releases.
// (e.g. cmake, quota)
type DepsProvisioner struct {
	fullStemcellCompatibility bool

	runner   boshsys.CmdRunner
	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewDepsProvisioner(
	fullStemcellCompatibility bool,
	runner boshsys.CmdRunner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) DepsProvisioner {
	return DepsProvisioner{
		fullStemcellCompatibility: fullStemcellCompatibility,

		runner:   runner,
		eventLog: eventLog,
		logger:   logger,
	}
}

func (p DepsProvisioner) Provision() error {
	pkgNames := depsProvisionerPkgsForMinimumStemcellCompatibility

	if p.fullStemcellCompatibility {
		pkgNames = append(pkgNames, depsProvisionerPkgsForFullStemcellCompatibility...)
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
	if strings.Contains(err.Error(), depsProvisionerAptGetUpdateMsg) {
		_, _, _, err := p.runner.RunCommand("apt-get", "-y", "update")
		if err != nil {
			return bosherr.WrapError(err, "Updating sources")
		}

		var lastInstallErr error

		// For some reason libssl-dev was really hard to install on the first try
		for i := 0; i < 3; i++ {
			_, _, _, lastInstallErr = p.runner.RunCommand("apt-get", "-y", "install", name)
			if lastInstallErr == nil {
				return nil
			}

			time.Sleep(1 * time.Second)
		}

		return bosherr.WrapError(lastInstallErr, "Installing %s after updating", name)
	}

	return err
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

var depsProvisionerPkgsForMinimumStemcellCompatibility = []string{
	// Most BOSH releases require it for packaging
	"build-essential", // 16sec
	"cmake",           // 6sec

	"libcap2-bin",
	"libcap2-dev",

	"libbz2-1.0",   // noop on precise64 Vagrant box
	"libbz2-dev",   // 2sec
	"libxslt1-dev", // 2sec
	"libxml2-dev",  // 2sec

	// Used by BOSH Agent
	"iputils-arping",

	// For warden
	"quota", // 1sec

	// Started needing that in saucy for building BOSH
	"libssl-dev",

	"bison",
	"flex",

	"gettext",
	"libreadline6-dev",
	"libncurses5-dev",
}

// Taken from base_apt stemcell builder stage
var depsProvisionerPkgsForFullStemcellCompatibility = []string{
	"libaio1",
	"uuid-dev",
	"nfs-common",
	"zlib1g-dev",
	"apparmor-utils",
	"openssh-server",

	"libgcrypt-dev",
	"ca-certificates",

	// CURL
	"libcurl3",
	"libcurl3-dev",

	// XML
	"libxml2",
	"libxml2-dev",
	"libxslt1.1",
	"libxslt1-dev",

	// Utils
	"bind9-host",
	"dnsutils",
	"zip",
	"unzip",
	"psmisc",
	"lsof",
	"strace",
	"curl",
	"wget",
	"gdb",
	"sysstat",
	"rsync",

	"iptables",
	"tcpdump",
	"traceroute",
}
