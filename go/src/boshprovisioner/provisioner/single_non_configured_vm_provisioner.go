package provisioner

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpeventlog "boshprovisioner/eventlog"
	bpvm "boshprovisioner/vm"
)

// SingleNonConfiguredVMProvisioner configures 1 VM as a regular empty BOSH VM.
type SingleNonConfiguredVMProvisioner struct {
	vmProvisioner bpvm.VMProvisioner
	eventLog      bpeventlog.Log
	logger        boshlog.Logger
}

func NewSingleNonConfiguredVMProvisioner(
	vmProvisioner bpvm.VMProvisioner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) SingleNonConfiguredVMProvisioner {
	return SingleNonConfiguredVMProvisioner{
		vmProvisioner: vmProvisioner,
		eventLog:      eventLog,
		logger:        logger,
	}
}

func (p SingleNonConfiguredVMProvisioner) Provision() error {
	// todo VM was possibly provisioned last time
	_, err := p.vmProvisioner.ProvisionNonConfigured()
	if err != nil {
		return bosherr.WrapError(err, "Provisioning VM")
	}

	// Do not Deprovision() VM to keep instance running

	return nil
}
