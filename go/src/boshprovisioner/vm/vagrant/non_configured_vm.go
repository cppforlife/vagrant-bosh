package vagrant

import (
	bpagclient "boshprovisioner/agent/client"
)

// NonConfiguredVM represents provisioned Vagrant machine
// that CANNOT be communicated with via an AgentClient.
type NonConfiguredVM struct {
	vmProvisioner *VMProvisioner
}

func NewNonConfiguredVM(vmProvisioner *VMProvisioner) NonConfiguredVM {
	return NonConfiguredVM{vmProvisioner: vmProvisioner}
}

func (vm NonConfiguredVM) AgentClient() bpagclient.Client {
	// Programmer error
	panic("Must not ask for AgentClient from a non-configured VM")
}

func (vm NonConfiguredVM) Deprovision() error {
	return vm.vmProvisioner.deprovision(vm)
}

func (vm NonConfiguredVM) vagrantVM() {}
