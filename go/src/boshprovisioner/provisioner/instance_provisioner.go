package provisioner

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpinstupd "boshprovisioner/instance/updater"
	bpvm "boshprovisioner/vm"
)

const instanceProvisionerLogTag = "InstanceProvisioner"

type InstanceProvisioner struct {
	instanceUpdaterFactory bpinstupd.UpdaterFactory
	vmProvisioner          bpvm.VMProvisioner
	logger                 boshlog.Logger
}

func NewInstanceProvisioner(
	instanceUpdaterFactory bpinstupd.UpdaterFactory,
	vmProvisioner bpvm.VMProvisioner,
	logger boshlog.Logger,
) InstanceProvisioner {
	return InstanceProvisioner{
		instanceUpdaterFactory: instanceUpdaterFactory,
		vmProvisioner:          vmProvisioner,
		logger:                 logger,
	}
}

func (p InstanceProvisioner) Provision(job bpdep.Job, instance bpdep.Instance) error {
	p.logger.Debug(instanceProvisionerLogTag, "Updating instance")

	vm, err := p.vmProvisioner.Provision(instance)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning agent")
	}

	updater := p.instanceUpdaterFactory.NewUpdater(vm.AgentClient(), job, instance)

	err = updater.Update()
	if err != nil {
		return bosherr.WrapError(err, "Updating instance %d", instance.Index)
	}

	return nil
}
