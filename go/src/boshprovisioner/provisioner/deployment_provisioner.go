package provisioner

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpeventlog "boshprovisioner/eventlog"
)

const deploymentProvisionerLogTag = "DeploymentProvisioner"

// DeploymentProvisioner interprets deployment manifest and
// configures system just like regular BOSH VM would be configured.
type DeploymentProvisioner struct {
	manifestPath            string
	deploymentReaderFactory bpdep.ReaderFactory

	releaseCompiler     ReleaseCompiler
	instanceProvisioner InstanceProvisioner

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewDeploymentProvisioner(
	manifestPath string,
	deploymentReaderFactory bpdep.ReaderFactory,
	releaseCompiler ReleaseCompiler,
	instanceProvisioner InstanceProvisioner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) DeploymentProvisioner {
	return DeploymentProvisioner{
		manifestPath:            manifestPath,
		deploymentReaderFactory: deploymentReaderFactory,

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

	job, instance, err := p.validateInstance(deployment)
	if task.End(err) != nil {
		return bosherr.WrapError(err, "Validating instance")
	}

	err = p.releaseCompiler.Compile(deployment.CompilationInstance, deployment.Releases)
	if err != nil {
		return bosherr.WrapError(err, "Compiling releases")
	}

	err = p.instanceProvisioner.Provision(job, instance)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning instance")
	}

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
