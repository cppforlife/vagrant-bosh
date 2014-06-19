package vm

import (
	bpagclient "boshprovisioner/agent/client"
)

// VM represents already provisioned machine
// that can be communicated with via an AgentClient.
type VM struct {
	agentClient bpagclient.Client
}

func NewVM(agentClient bpagclient.Client) VM {
	return VM{agentClient: agentClient}
}

func (vm VM) AgentClient() bpagclient.Client {
	return vm.agentClient
}
