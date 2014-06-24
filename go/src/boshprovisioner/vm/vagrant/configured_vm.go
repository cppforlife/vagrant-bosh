package vagrant

import (
	bpagclient "boshprovisioner/agent/client"
)

// ConfiguredVM represents provisioned Vagrant machine
// that can be communicated with via an AgentClient.
type ConfiguredVM struct {
	vmProvisioner *VMProvisioner
	agentClient   bpagclient.Client
}

func NewConfiguredVM(vmProvisioner *VMProvisioner, agentClient bpagclient.Client) ConfiguredVM {
	return ConfiguredVM{
		vmProvisioner: vmProvisioner,
		agentClient:   agentClient,
	}
}

func (vm ConfiguredVM) AgentClient() bpagclient.Client {
	return vm.agentClient
}

func (vm ConfiguredVM) Deprovision() error {
	return vm.vmProvisioner.deprovision(vm)
}

func (vm ConfiguredVM) vagrantVM() {}
