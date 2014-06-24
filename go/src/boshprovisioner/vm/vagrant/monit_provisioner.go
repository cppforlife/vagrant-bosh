package vagrant

import (
	"time"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
)

const monitProvisionerLogTag = "MonitProvisioner"

// MonitProvisioner installs Monit binary.
type MonitProvisioner struct {
	cmds             SimpleCmds
	assetManager     AssetManager
	runitProvisioner RunitProvisioner
	logger           boshlog.Logger
}

func NewMonitProvisioner(
	cmds SimpleCmds,
	assetManager AssetManager,
	runitProvisioner RunitProvisioner,
	logger boshlog.Logger,
) MonitProvisioner {
	return MonitProvisioner{
		cmds:             cmds,
		assetManager:     assetManager,
		runitProvisioner: runitProvisioner,
		logger:           logger,
	}
}

func (p MonitProvisioner) Provision() error {
	path := "/var/vcap/monit"

	err := p.cmds.MkdirP(path)
	if err != nil {
		return err
	}

	err = p.configureMonitrc()
	if err != nil {
		return bosherr.WrapError(err, "Configuring monitrc")
	}

	err = p.runitProvisioner.Provision("monit", 1*time.Minute)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning monit with runit")
	}

	return nil
}

func (p MonitProvisioner) configureMonitrc() error {
	p.logger.Info(monitProvisionerLogTag, "Configuring monitc")

	err := p.cmds.MkdirP("/var/vcap/bosh/etc")
	if err != nil {
		return err
	}

	err = p.assetManager.Place("monit/monitrc", "/var/vcap/bosh/etc/monitrc")
	if err != nil {
		return bosherr.WrapError(err, "Placing monitrc")
	}

	err = p.cmds.Chmod("700", "/var/vcap/bosh/etc/monitrc")
	if err != nil {
		return err
	}

	// monit refuses to start without an include file present
	err = p.cmds.MkdirP("/var/vcap/monit")
	if err != nil {
		return err
	}

	err = p.cmds.Touch("/var/vcap/monit/empty.monitrc")
	if err != nil {
		return err
	}

	return nil
}
