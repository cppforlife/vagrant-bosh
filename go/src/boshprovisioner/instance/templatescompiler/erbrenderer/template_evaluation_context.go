package erbrenderer

import (
	"encoding/json"

	bosherr "bosh/errors"

	bpdep "boshprovisioner/deployment"
	bpreljob "boshprovisioner/release/job"
)

type TemplateEvaluationContext struct {
	relJob   bpreljob.Job
	instance bpdep.Instance
}

// rootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type rootContext struct {
	Index int `json:"index"`

	JobContext jobContext `json:"job"`

	Deployment string `json:"deployment"`

	// Usually is accessed with <%= spec.networks.default.ip %>
	NetworkContexts map[string]networkContext `json:"networks"`

	Properties map[string]interface{} `json:"properties"`
}

type jobContext struct {
	Name string `json:"name"`
}

// networkContext is not fully backwards compatible.
type networkContext struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

func NewTemplateEvaluationContext(
	relJob bpreljob.Job,
	instance bpdep.Instance,
) TemplateEvaluationContext {
	return TemplateEvaluationContext{
		relJob:   relJob,
		instance: instance,
	}
}

func (c TemplateEvaluationContext) MarshalJSON() ([]byte, error) {
	properties, err := NewRenderProperties(c.relJob, c.instance).AsMap()
	if err != nil {
		return nil, bosherr.WrapError(err, "Rendering properties")
	}

	context := rootContext{
		Index:      c.instance.Index,
		JobContext: jobContext{Name: c.instance.JobName},
		Deployment: c.instance.DeploymentName,

		NetworkContexts: c.buildNetworkContexts(),
		Properties:      properties,
	}

	return json.Marshal(context)
}

func (c TemplateEvaluationContext) buildNetworkContexts() map[string]networkContext {
	networkContexts := map[string]networkContext{}

	for _, na := range c.instance.NetworkAssociations {
		netConfig := c.instance.NetworkConfigurationForNetworkAssociation(na)

		networkContexts[na.Network.Name] = networkContext{
			IP:      netConfig.IP,
			Netmask: netConfig.Netmask,
			Gateway: netConfig.Gateway,
		}
	}

	return networkContexts
}
