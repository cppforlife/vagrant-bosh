package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
)

const vmProvisionerLogTag = "VMProvisioner"

// VMProvisioner installs system dependencies that
// are usually found on a stemcell, adds vcap user,
// and finally install Agent and Monit.
type VMProvisioner struct {
	vcapUserProvisioner VCAPUserProvisioner
	depsProvisioner     DepsProvisioner
	agentProvisioner    AgentProvisioner

	logger boshlog.Logger
}

func NewVMProvisioner(
	vcapUserProvisioner VCAPUserProvisioner,
	depsProvisioner DepsProvisioner,
	agentProvisioner AgentProvisioner,
	logger boshlog.Logger,
) VMProvisioner {
	return VMProvisioner{
		vcapUserProvisioner: vcapUserProvisioner,
		depsProvisioner:     depsProvisioner,
		agentProvisioner:    agentProvisioner,

		logger: logger,
	}
}

func (p VMProvisioner) Provision(instance bpdep.Instance) (VM, error) {
	var vm VM

	err := p.vcapUserProvisioner.Provision()
	if err != nil {
		return vm, bosherr.WrapError(err, "Provisioning vcap user")
	}

	err = p.depsProvisioner.Provision()
	if err != nil {
		return vm, bosherr.WrapError(err, "Provisioning dependencies")
	}

	agentClient, err := p.agentProvisioner.Provision(instance)
	if err != nil {
		return vm, bosherr.WrapError(err, "Provisioning agent")
	}

	return NewVM(agentClient), nil
}
