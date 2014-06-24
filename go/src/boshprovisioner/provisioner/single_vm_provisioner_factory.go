package provisioner

import (
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpeventlog "boshprovisioner/eventlog"
	bpinstance "boshprovisioner/instance"
	bpvm "boshprovisioner/vm"
)

type SingleVMProvisionerFactory struct {
	deploymentReaderFactory     bpdep.ReaderFactory
	deploymentProvisionerConfig DeploymentProvisionerConfig

	vmProvisioner       bpvm.VMProvisioner
	releaseCompiler     ReleaseCompiler
	instanceProvisioner bpinstance.InstanceProvisioner

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewSingleVMProvisionerFactory(
	deploymentReaderFactory bpdep.ReaderFactory,
	deploymentProvisionerConfig DeploymentProvisionerConfig,
	vmProvisioner bpvm.VMProvisioner,
	releaseCompiler ReleaseCompiler,
	instanceProvisioner bpinstance.InstanceProvisioner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) SingleVMProvisionerFactory {
	return SingleVMProvisionerFactory{
		deploymentReaderFactory:     deploymentReaderFactory,
		deploymentProvisionerConfig: deploymentProvisionerConfig,

		vmProvisioner:       vmProvisioner,
		releaseCompiler:     releaseCompiler,
		instanceProvisioner: instanceProvisioner,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (f SingleVMProvisionerFactory) NewSingleVMProvisioner() DeploymentProvisioner {
	var prov DeploymentProvisioner

	if len(f.deploymentProvisionerConfig.ManifestPath) > 0 {
		prov = NewSingleConfiguredVMProvisioner(
			f.deploymentProvisionerConfig.ManifestPath,
			f.deploymentReaderFactory,
			f.vmProvisioner,
			f.releaseCompiler,
			f.instanceProvisioner,
			f.eventLog,
			f.logger,
		)
	} else {
		prov = NewSingleNonConfiguredVMProvisioner(
			f.vmProvisioner,
			f.eventLog,
			f.logger,
		)
	}

	return prov
}
