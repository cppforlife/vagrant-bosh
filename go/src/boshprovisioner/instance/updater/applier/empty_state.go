package applier

import (
	boshas "bosh/agent/applier/applyspec"

	bpdep "boshprovisioner/deployment"
)

// EmptyState represents state for a VM
// that should not be running any job templates.
type EmptyState struct {
	instance bpdep.Instance
}

func NewEmptyState(instance bpdep.Instance) EmptyState {
	return EmptyState{instance: instance}
}

func (s EmptyState) AsApplySpec() boshas.V1ApplySpec {
	var spec boshas.V1ApplySpec

	jobName := s.instance.JobName
	jobIndex := s.instance.Index

	spec = boshas.V1ApplySpec{
		ConfigurationHash: "fake-configuration-hash", // todo

		Deployment: s.instance.DeploymentName,

		JobSpec: boshas.JobSpec{
			Name: &jobName,
		},

		Index: &jobIndex,

		NetworkSpecs: s.buildNetworkSpecs(),

		// todo find out whats here
		ResourcePoolSpecs: s.buildResourcePoolSpecs(),
	}

	return spec
}

func (s EmptyState) buildNetworkSpecs() map[string]boshas.NetworkSpec {
	specs := map[string]boshas.NetworkSpec{}

	for _, netAssoc := range s.instance.NetworkAssociations {
		netConfig := s.instance.NetworkConfigurationForNetworkAssociation(netAssoc)

		specs[netAssoc.Network.Name] = boshas.NetworkSpec{
			Fields: map[string]interface{}{
				"type":    netAssoc.Network.Type,
				"ip":      netConfig.IP,
				"netmask": netConfig.Netmask,
				"gateway": netConfig.Gateway,
			},
		}
	}

	return specs
}

func (s EmptyState) buildResourcePoolSpecs() map[string]interface{} {
	return map[string]interface{}{}
}
