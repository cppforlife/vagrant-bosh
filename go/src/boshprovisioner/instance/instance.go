package instance

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpinstupd "boshprovisioner/instance/updater"
)

type Instance struct {
	updater bpinstupd.Updater

	job         bpdep.Job
	depInstance bpdep.Instance

	logger boshlog.Logger
}

func NewInstance(
	updater bpinstupd.Updater,
	job bpdep.Job,
	depInstance bpdep.Instance,
	logger boshlog.Logger,
) Instance {
	return Instance{
		updater:     updater,
		job:         job,
		depInstance: depInstance,
		logger:      logger,
	}
}

func (i Instance) Deprovision() error {
	i.logger.Debug(instanceProvisionerLogTag, "Tearing down instance")

	err := i.updater.TearDown()
	if err != nil {
		return bosherr.WrapError(err, "Tearing down instance %d", i.depInstance.Index)
	}

	return nil
}
