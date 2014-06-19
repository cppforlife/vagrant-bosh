package deployment

import (
	"fmt"
)

type NetworkConfiguration struct {
	IP      string
	Netmask string
	Gateway string
}

func (i Instance) NetworkConfigurationForNetworkAssociation(na NetworkAssociation) NetworkConfiguration {
	// If instance is always has static ip assigned there is nothing to resolve
	if na.MustHaveStaticIP {
		return NetworkConfiguration{IP: na.StaticIP.String()}
	}

	// Dynamic network information can be resolved after VM is powered on.
	if na.Network.Type == NetworkTypeDynamic {
		networkSpec, ok := i.CurrentState.NetworkSpecs[na.Network.Name]
		if !ok {
			return NetworkConfiguration{}
		}

		// todo better way to deserialize? structmap?
		ipStr, ok := networkSpec.Fields["ip"].(string)
		if !ok {
			return NetworkConfiguration{}
		}

		netmaskStr, ok := networkSpec.Fields["netmask"].(string)
		if !ok {
			return NetworkConfiguration{}
		}

		gatewayStr, ok := networkSpec.Fields["gateway"].(string)
		if !ok {
			return NetworkConfiguration{}
		}

		return NetworkConfiguration{
			IP:      ipStr,
			Netmask: netmaskStr,
			Gateway: gatewayStr,
		}
	}

	return NetworkConfiguration{}
}

func (i Instance) DNDRecordName(na NetworkAssociation) string {
	return fmt.Sprintf("%d.%s.%s.%s.bosh", i.Index, i.JobName, na.Network.Name, i.DeploymentName)
}
