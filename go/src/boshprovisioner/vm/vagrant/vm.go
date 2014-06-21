package vm

import (
	bpagclient "boshprovisioner/agent/client"
)

// Vagrant VM represents already provisioned Vagrant machine
// that can be communicated with via an AgentClient.
type VM struct {
	vmProvisioner *VMProvisioner
	agentClient   bpagclient.Client
}

func NewVM(vmProvisioner *VMProvisioner, agentClient bpagclient.Client) VM {
	return VM{
		vmProvisioner: vmProvisioner,
		agentClient:   agentClient,
	}
}

func (vm VM) AgentClient() bpagclient.Client {
	return vm.agentClient
}

func (vm VM) Deprovision() error {
	return vm.vmProvisioner.Deprovision(vm)
}
