package provisioner

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpeventlog "boshprovisioner/eventlog"
	bpinstance "boshprovisioner/instance"
	bpvm "boshprovisioner/vm"
)

const deploymentProvisionerLogTag = "DeploymentProvisioner"

// DeploymentProvisioner interprets deployment manifest and
// configures system just like regular BOSH VM would be configured.
type DeploymentProvisioner struct {
	manifestPath            string
	deploymentReaderFactory bpdep.ReaderFactory

	vmProvisioner       bpvm.VMProvisioner
	releaseCompiler     ReleaseCompiler
	instanceProvisioner bpinstance.InstanceProvisioner

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewDeploymentProvisioner(
	manifestPath string,
	deploymentReaderFactory bpdep.ReaderFactory,
	vmProvisioner bpvm.VMProvisioner,
	releaseCompiler ReleaseCompiler,
	instanceProvisioner bpinstance.InstanceProvisioner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) DeploymentProvisioner {
	return DeploymentProvisioner{
		manifestPath:            manifestPath,
		deploymentReaderFactory: deploymentReaderFactory,

		vmProvisioner:       vmProvisioner,
		releaseCompiler:     releaseCompiler,
		instanceProvisioner: instanceProvisioner,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (p DeploymentProvisioner) Provision() error {
	stage := p.eventLog.BeginStage("Setting up instance", 2)

	reader := p.deploymentReaderFactory.NewManifestReader(p.manifestPath)

	task := stage.BeginTask("Reading deployment manifest")

	deployment, err := reader.Read()
	if task.End(err) != nil {
		return bosherr.WrapError(err, "Reading deployment")
	}

	task = stage.BeginTask("Validating instance")

	job, depInstance, err := p.validateInstance(deployment)
	if task.End(err) != nil {
		return bosherr.WrapError(err, "Validating instance")
	}

	// todo VM was possibly provisioned last time
	vm, err := p.vmProvisioner.Provision(depInstance)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning VM")
	}

	instance := p.instanceProvisioner.PreviouslyProvisioned(vm.AgentClient(), job, depInstance)

	err = instance.Deprovision()
	if err != nil {
		return bosherr.WrapError(err, "Deprovisioning instance")
	}

	// Deprovision VM before using release compiler since it will try to provision its own VM
	err = vm.Deprovision()
	if err != nil {
		return bosherr.WrapError(err, "Deprovisioning VM")
	}

	err = p.releaseCompiler.Compile(deployment.CompilationInstance, deployment.Releases)
	if err != nil {
		return bosherr.WrapError(err, "Compiling releases")
	}

	vm, err = p.vmProvisioner.Provision(depInstance)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning VM")
	}

	_, err = p.instanceProvisioner.Provision(vm.AgentClient(), job, depInstance)
	if err != nil {
		return bosherr.WrapError(err, "Starting instance")
	}

	// Do not Deprovision() VM to keep instance running

	return nil
}

func (p DeploymentProvisioner) validateInstance(deployment bpdep.Deployment) (bpdep.Job, bpdep.Instance, error) {
	p.logger.Debug(deploymentProvisionerLogTag, "Validate instance")

	var job bpdep.Job
	var instance bpdep.Instance

	if len(deployment.Jobs) > 1 {
		return job, instance, bosherr.New("Must have exactly 1 job")
	}

	job = deployment.Jobs[0]

	if len(job.Instances) != 1 {
		return job, instance, bosherr.New("Must have exactly 1 instance")
	}

	instance = job.Instances[0]

	return job, instance, nil
}
