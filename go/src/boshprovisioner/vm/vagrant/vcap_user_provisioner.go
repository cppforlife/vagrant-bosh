package vagrant

import (
	"strings"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpeventlog "boshprovisioner/eventlog"
)

const vcapUserProvisionerLogTag = "VCAPUserProvisioner"

// VCAPUserProvisioner adds and configures vcap user.
type VCAPUserProvisioner struct {
	fs       boshsys.FileSystem
	runner   boshsys.CmdRunner
	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewVCAPUserProvisioner(
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) VCAPUserProvisioner {
	return VCAPUserProvisioner{
		fs:       fs,
		runner:   runner,
		eventLog: eventLog,
		logger:   logger,
	}
}

func (p VCAPUserProvisioner) Provision() error {
	stage := p.eventLog.BeginStage("Setting up vcap user", 3)

	task := stage.BeginTask("Adding vcap user")

	err := task.End(p.setUpVcapUser())
	if err != nil {
		return bosherr.WrapError(err, "Setting up vcap user")
	}

	task = stage.BeginTask("Configuring locales")

	err = task.End(p.configureLocales())
	if err != nil {
		return bosherr.WrapError(err, "Configuring locales")
	}

	task = stage.BeginTask("Harden permissions")

	err = task.End(p.hardenPermissinons())
	if err != nil {
		return bosherr.WrapError(err, "Harden permissions")
	}

	return nil
}

func (p VCAPUserProvisioner) setUpVcapUser() error {
	p.logger.Info(vcapUserProvisionerLogTag, "Setting up vcap user")

	_, stderr, _, err := p.runner.RunCommand("groupadd", "--system", "admin")
	if err != nil {
		if !strings.Contains(stderr, "group 'admin' already exists") {
			return err
		}
	}

	_, stderr, _, err = p.runner.RunCommand("useradd", "-m", "--comment", "BOSH System User", "vcap")
	if err != nil {
		if !strings.Contains(stderr, "user 'vcap' already exists") {
			return err
		}
	}

	cmds := [][]string{
		{"bash", "-c", "echo 'vcap:c1oudc0w' | chpasswd"},
		{"bash", "-c", "echo 'root:c1oudc0w' | chpasswd"},

		{"usermod", "-G", "admin,adm,audio,cdrom,dialout,floppy,video,dip,plugdev", "vcap"},
		{"usermod", "-s", "/bin/bash", "vcap"},
	}

	for _, cmd := range cmds {
		_, _, _, err := p.runner.RunCommand(cmd[0], cmd[1:]...)
		if err != nil {
			return err
		}
	}

	// todo setup vcap no-password sudo access

	_, stderr, _, err = p.runner.RunCommand("usermod", "-a", "-G", "vcap", "vagrant")
	if err != nil {
		if !strings.Contains(stderr, "user 'vagrant' does not exist") {
			return err
		}
	}

	return p.setUpBoshBinPath()
}

func (p VCAPUserProvisioner) setUpBoshBinPath() error {
	boshBinExport := "export PATH=/var/vcap/bosh/bin:$PATH"

	for _, bashrcPath := range []string{"/root/.bashrc", "/home/vcap/.bashrc"} {
		contents, err := p.fs.ReadFileString(bashrcPath)
		if err != nil {
			return err
		}

		if !strings.Contains(contents, boshBinExport) {
			err := p.fs.WriteFileString(bashrcPath, contents+"\n"+boshBinExport+"\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p VCAPUserProvisioner) configureLocales() error {
	_, _, _, err := p.runner.RunCommand("locale-gen", "en_US.UTF-8")
	if err != nil {
		return err
	}

	_, _, _, err = p.runner.RunCommand("dpkg-reconfigure", "locales")
	if err != nil {
		return err
	}

	// Configure vcap user locale (postgres initdb fails if mismatched)
	return p.fs.WriteFileString("/etc/default/locale", "LANG=en_US.UTF-8\nLC_ALL=en_US.UTF-8")
}

func (p VCAPUserProvisioner) hardenPermissinons() error {
	cmds := [][]string{
		{"bash", "-c", "echo 'vcap' > /etc/cron.allow"},
		{"bash", "-c", "echo 'vcap' > /etc/at.allow"},

		{"chmod", "0770", "/var/lock"},
		{"chown", "-h", "root:vcap", "/var/lock"},
		{"chown", "-LR", "root:vcap", "/var/lock"},

		{"chmod", "0640", "/etc/cron.allow"},
		{"chown", "root:vcap", "/etc/cron.allow"},

		{"chmod", "0640", "/etc/at.allow"},
		{"chown", "root:vcap", "/etc/at.allow"},
	}

	for _, cmd := range cmds {
		_, _, _, err := p.runner.RunCommand(cmd[0], cmd[1:]...)
		if err != nil {
			return err
		}
	}

	return nil
}
