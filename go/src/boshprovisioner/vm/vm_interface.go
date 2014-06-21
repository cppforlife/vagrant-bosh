package vm

import (
	bpagclient "boshprovisioner/agent/client"
	bpdep "boshprovisioner/deployment"
)

type VMProvisioner interface {
	// todo should not rely on bpdep.Instance
	Provision(bpdep.Instance) (VM, error)
}

type VM interface {
	AgentClient() bpagclient.Client

	Deprovision() error
}
