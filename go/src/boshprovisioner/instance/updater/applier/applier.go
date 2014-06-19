package applier

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
	bpdep "boshprovisioner/deployment"
	bptplcomp "boshprovisioner/instance/templatescompiler"
	bppkgscomp "boshprovisioner/packagescompiler"
)

const applierLogTag = "Applier"

type Applier struct {
	depJob   bpdep.Job
	instance bpdep.Instance

	templatesCompiler bptplcomp.TemplatesCompiler
	packagesCompiler  bppkgscomp.PackagesCompiler

	agentClient bpagclient.Client
	logger      boshlog.Logger
}

func NewApplier(
	depJob bpdep.Job,
	instance bpdep.Instance,
	templatesCompiler bptplcomp.TemplatesCompiler,
	packagesCompiler bppkgscomp.PackagesCompiler,
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) Applier {
	return Applier{
		depJob:   depJob,
		instance: instance,

		templatesCompiler: templatesCompiler,
		packagesCompiler:  packagesCompiler,

		agentClient: agentClient,
		logger:      logger,
	}
}

func (a Applier) Apply() error {
	a.logger.Debug(applierLogTag, "Applying empty state")

	emptyState := NewEmptyState(a.instance)

	_, err := a.agentClient.Apply(emptyState.AsApplySpec())
	if err != nil {
		return bosherr.WrapError(err, "Applying empty spec")
	}

	// Changes local copy of an instance
	a.instance.CurrentState, err = a.agentClient.GetState()
	if err != nil {
		return bosherr.WrapError(err, "Getting state")
	}

	a.logger.Debug(applierLogTag, "Finished applying empty state")

	// Recompile job templates since current instance state might have changed.
	// e.g. dynamic IP could now be set
	err = a.templatesCompiler.Compile(a.depJob, a.instance)
	if err != nil {
		return bosherr.WrapError(err, "Compiling templates %s", a.depJob.Name)
	}

	a.logger.Debug(applierLogTag, "Applying job state")

	jobState := NewJobState(
		a.depJob,
		a.instance,
		a.templatesCompiler,
		a.packagesCompiler,
	)

	jobApplySpec, err := jobState.AsApplySpec()
	if err != nil {
		return err
	}

	_, err = a.agentClient.Apply(jobApplySpec)
	if err != nil {
		return bosherr.WrapError(err, "Applying job spec")
	}

	a.logger.Debug(applierLogTag, "Finished applying job state")

	return nil
}
