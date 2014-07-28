package vm

import (
	bpagclient "boshprovisioner/agent/client"
	bpdep "boshprovisioner/deployment"
)

type Provisioner interface {
	// Provision creates and configures VM for future agent communication.
	// todo should not rely on bpdep.Instance
	Provision(bpdep.Instance) (VM, error)

	// ProvisionNonConfigured creates and does NOT configure VM for communication.
	ProvisionNonConfigured() (VM, error)
}

type VM interface {
	// AgentClient returns a client immediately ready for communication.
	AgentClient() bpagclient.Client

	// Deprovision deletes VM previously provisioned VM.
	Deprovision() error
}

type ProvisionerConfig struct {
	// When provisioning, install all dependencies that official stemcells carry.
	// By default, provisioners will only install absolutely needed dependencies.
	FullStemcellCompatibility bool `json:"full_stemcell_compatibility"`

	AgentProvisioner AgentProvisionerConfig `json:"agent_provisioner"`
}

type AgentProvisionerConfig struct {
	// e.g. warden, aws
	Infrastructure string `json:"infrastructure"`

	// e.g. ubuntu, centos
	Platform string `json:"platform"`

	// Usually save to /var/vcap/bosh/agent.json
	Configuration map[string]interface{} `json:"configuration"`

	// e.g. "https://user:password@127.0.0.1:4321/agent"
	Mbus string `json:"mbus"`
}
