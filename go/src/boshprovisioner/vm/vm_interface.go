package vm

import (
	bpagclient "boshprovisioner/agent/client"
	bpdep "boshprovisioner/deployment"
)

type VMProvisioner interface {
	// Provision creates and configures VM to be usable by BOSH releases.
	// todo should not rely on bpdep.Instance
	Provision(bpdep.Instance) (VM, error)
}

type VMProvisionerConfig struct {
	// When provisioning, install all dependencies that official stemcells carry.
	// By default, provisioners will only install absolutely needed dependencies.
	FullStemcellCompatibility bool `json:"full_stemcell_compatibility"`
}

type VM interface {
	// AgentClient returns a client immediately ready for communication.
	AgentClient() bpagclient.Client

	// Deprovision deletes VM previously provisioned VM.
	Deprovision() error
}
