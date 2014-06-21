package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpeventlog "boshprovisioner/eventlog"
)

const vcapUserProvisionerLogTag = "VCAPUserProvisioner"

// VCAPUserProvisioner adds and configures vcap user.
type VCAPUserProvisioner struct {
	cmds     SimpleCmds
	runner   boshsys.CmdRunner
	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewVCAPUserProvisioner(
	cmds SimpleCmds,
	runner boshsys.CmdRunner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) VCAPUserProvisioner {
	return VCAPUserProvisioner{
		cmds:     cmds,
		runner:   runner,
		eventLog: eventLog,
		logger:   logger,
	}
}

func (p VCAPUserProvisioner) Provision() error {
	stage := p.eventLog.BeginStage("Setting up vcap user", 2)

	task := stage.BeginTask("Adding vcap user")

	err := task.End(p.setUpVcapUser())
	if err != nil {
		return bosherr.WrapError(err, "Setting up vcap user")
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

	userBash := `
    groupadd --system admin
    useradd -m --comment 'BOSH System User' vcap

    echo "vcap:c1oudc0w" | chpasswd
    echo "root:c1oudc0w" | chpasswd

    usermod -G admin,adm,audio,cdrom,dialout,floppy,video,dip,plugdev vcap
    usermod -s /bin/bash vcap
  `

	err := p.cmds.Bash(userBash)
	if err != nil {
		return err
	}

	// todo setup vcap no-password sudo access
	_, _, _, err = p.runner.RunCommand("usermod", "-a", "-G", "vcap", "vagrant")
	if err != nil {
		return err
	}

	envBashs := []string{
		"echo 'export PATH=/var/vcap/bosh/bin:$PATH' >> /root/.bashrc",
		"echo 'export PATH=/var/vcap/bosh/bin:$PATH' >> /home/vcap/.bashrc",

		// Configure vcap user locale (postgres initdb fails if mismatched)
		"echo 'LANG=en_US.UTF-8\nLC_ALL=en_US.UTF-8' > /etc/default/locale",
	}

	for _, bash := range envBashs {
		err := p.cmds.Bash(bash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p VCAPUserProvisioner) hardenPermissinons() error {
	permsBash := `
    echo 'vcap' > /etc/cron.allow
    echo 'vcap' > /etc/at.allow

    chmod 0770 /var/lock
    chown -h root:vcap /var/lock
    chown -LR root:vcap /var/lock

    chmod 0640 /etc/cron.allow
    chown root:vcap /etc/cron.allow

    chmod 0640 /etc/at.allow
    chown root:vcap /etc/at.allow
  `

	err := p.cmds.Bash(permsBash)
	if err != nil {
		return err
	}

	return nil
}
