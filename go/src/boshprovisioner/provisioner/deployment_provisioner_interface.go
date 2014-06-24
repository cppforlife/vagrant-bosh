package provisioner

type DeploymentProvisioner interface {
	Provision() error
}

type DeploymentProvisionerConfig struct {
	// If manifest path is empty, release compilation and job provisioning will be skipped
	ManifestPath string `json:"manifest_path"`
}
