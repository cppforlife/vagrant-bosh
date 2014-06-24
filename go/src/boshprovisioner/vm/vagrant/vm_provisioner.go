package vagrant

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpvm "boshprovisioner/vm"
)

var (
	ErrAlreadyProvisioned = bosherr.New("Vagrant VM is already provisioned")
	ErrNotProvisioned     = bosherr.New("Vagrant VM is not provisioned")
)

// VMProvisioner installs system dependencies that are usually
// found on a stemcell, adds vcap user and finally install Agent and Monit.
type VMProvisioner struct {
	vcapUserProvisioner VCAPUserProvisioner
	depsProvisioner     DepsProvisioner
	agentProvisioner    AgentProvisioner

	// Cannot provision more than 1 VM at a time until previous VM is deprovisioned.
	vmProvisioned bool

	// Remember if we have recently run provisioning for a VM to skip it next time.
	vmPreviouslyProvisioned bool

	logger boshlog.Logger
}

type vagrantVM interface {
	vagrantVM()
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
	if p.vmProvisioned {
		return nil, ErrAlreadyProvisioned
	}

	err := p.provisionNonConfigured()
	if err != nil {
		return nil, err
	}

	agentClient, err := p.agentProvisioner.Configure(instance)
	if err != nil {
		return nil, bosherr.WrapError(err, "Configuring agent")
	}

	// Set this as late as possible so that caller can retry failed provisioning attempts.
	p.vmProvisioned = true

	return NewConfiguredVM(p, agentClient), nil
}

func (p *VMProvisioner) ProvisionNonConfigured() (bpvm.VM, error) {
	if p.vmProvisioned {
		return nil, ErrAlreadyProvisioned
	}

	err := p.provisionNonConfigured()
	if err != nil {
		return nil, err
	}

	// Set this as late as possible so that caller can retry failed provisioning attempts.
	p.vmProvisioned = true

	return NewNonConfiguredVM(p), nil
}

func (p *VMProvisioner) deprovision(vm vagrantVM) error {
	if !p.vmProvisioned {
		return ErrNotProvisioned
	}

	p.vmProvisioned = false

	return nil
}

func (p *VMProvisioner) provisionNonConfigured() error {
	if !p.vmPreviouslyProvisioned {
		err := p.vcapUserProvisioner.Provision()
		if err != nil {
			return bosherr.WrapError(err, "Provisioning vcap user")
		}

		err = p.depsProvisioner.Provision()
		if err != nil {
			return bosherr.WrapError(err, "Provisioning dependencies")
		}
	}

	p.vmPreviouslyProvisioned = true

	err := p.agentProvisioner.Provision()
	if err != nil {
		return bosherr.WrapError(err, "Provisioning agent")
	}

	return nil
}
