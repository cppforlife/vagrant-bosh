package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpvm "boshprovisioner/vm"
)

// Vagrant's VMProvisioner installs system dependencies that are usually
// found on a stemcell, adds vcap user and finally install Agent and Monit.
type VMProvisioner struct {
	vcapUserProvisioner VCAPUserProvisioner
	depsProvisioner     DepsProvisioner
	agentProvisioner    AgentProvisioner

	// Cannot provision more than 1 VM at a time until previous VM is deprovisioned.
	vmProvisioned bool

	// Remember if we have recently run provisioning for a VM to skip it next time
	vmPreviouslyProvisioned bool

	logger boshlog.Logger
}

func NewVMProvisioner(
	vcapUserProvisioner VCAPUserProvisioner,
	depsProvisioner DepsProvisioner,
	agentProvisioner AgentProvisioner,
	logger boshlog.Logger,
) *VMProvisioner {
	return &VMProvisioner{
		vcapUserProvisioner: vcapUserProvisioner,
		depsProvisioner:     depsProvisioner,
		agentProvisioner:    agentProvisioner,

		logger: logger,
	}
}

func (p *VMProvisioner) Provision(instance bpdep.Instance) (bpvm.VM, error) {
	var vm VM

	if p.vmProvisioned {
		return vm, bosherr.New("Vagrant VM is already provisioned")
	}

	if !p.vmPreviouslyProvisioned {
		err := p.vcapUserProvisioner.Provision()
		if err != nil {
			return vm, bosherr.WrapError(err, "Provisioning vcap user")
		}

		err = p.depsProvisioner.Provision()
		if err != nil {
			return vm, bosherr.WrapError(err, "Provisioning dependencies")
		}
	}

	p.vmPreviouslyProvisioned = true

	agentClient, err := p.agentProvisioner.Provision(instance)
	if err != nil {
		return vm, bosherr.WrapError(err, "Provisioning agent")
	}

	p.vmProvisioned = true

	return NewVM(p, agentClient), nil
}

func (p *VMProvisioner) Deprovision(vm VM) error {
	if !p.vmProvisioned {
		return bosherr.New("Vagrant VM is not provisioned")
	}

	p.vmProvisioned = false

	return nil
}
