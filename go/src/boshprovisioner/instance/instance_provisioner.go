package instance

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
	bpdep "boshprovisioner/deployment"
	bpinstupd "boshprovisioner/instance/updater"
)

const instanceProvisionerLogTag = "InstanceProvisioner"

type InstanceProvisioner struct {
	instanceUpdaterFactory bpinstupd.UpdaterFactory
	logger                 boshlog.Logger
}

func NewInstanceProvisioner(
	instanceUpdaterFactory bpinstupd.UpdaterFactory,
	logger boshlog.Logger,
) InstanceProvisioner {
	return InstanceProvisioner{
		instanceUpdaterFactory: instanceUpdaterFactory,
		logger:                 logger,
	}
}

func (p InstanceProvisioner) Provision(ac bpagclient.Client, job bpdep.Job, depInstance bpdep.Instance) (Instance, error) {
	p.logger.Debug(instanceProvisionerLogTag, "Updating instance")

	updater := p.instanceUpdaterFactory.NewUpdater(ac, job, depInstance)

	err := updater.SetUp()
	if err != nil {
		return Instance{}, bosherr.WrapError(err, "Updating instance %d", depInstance.Index)
	}

	return NewInstance(updater, job, depInstance, p.logger), nil
}

func (p InstanceProvisioner) PreviouslyProvisioned(ac bpagclient.Client, job bpdep.Job, depInstance bpdep.Instance) Instance {
	p.logger.Debug(instanceProvisionerLogTag, "Finding previously provisioned instance")

	updater := p.instanceUpdaterFactory.NewUpdater(ac, job, depInstance)

	return NewInstance(updater, job, depInstance, p.logger)
}
