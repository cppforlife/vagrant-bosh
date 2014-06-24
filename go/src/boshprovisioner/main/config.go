package main

import (
	"encoding/json"

	bosherr "bosh/errors"
	boshsys "bosh/system"

	bpprov "boshprovisioner/provisioner"
	bpvm "boshprovisioner/vm"
)

var DefaultWardenConfig = Config{
	VMProvisioner: bpvm.VMProvisionerConfig{
		AgentProvisioner: bpvm.AgentProvisionerConfig{
			Infrastructure: "warden",
			Platform:       "ubuntu",
			Mbus:           "https://user:password@127.0.0.1:4321/agent",
		},
	},
}

var DefaultWardenAgentConfiguration = map[string]interface{}{
	"Platform": map[string]interface{}{
		"Linux": map[string]interface{}{
			"UseDefaultTmpDir":              true,
			"UsePreformattedPersistentDisk": true,
			"BindMountPersistentDisk":       true,
		},
	},
}

type Config struct {
	// Assets dir is used as a temporary location to transfer files from host to guest.
	// It will not be created since assets already must be present.
	AssetsDir string `json:"assets_dir"`

	// Repos dir is mainly used to record what was placed in the blobstore.
	// It will be created if it does not exist.
	ReposDir string `json:"repos_dir"`

	Blobstore bpprov.BlobstoreConfig `json:"blobstore"`

	VMProvisioner bpvm.VMProvisionerConfig `json:"vm_provisioner"`

	DeploymentProvisioner bpprov.DeploymentProvisionerConfig `json:"deployment_provisioner"`
}

func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapError(err, "Reading config %s", path)
	}

	config = DefaultWardenConfig

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	if config.VMProvisioner.AgentProvisioner.Configuration == nil {
		config.VMProvisioner.AgentProvisioner.Configuration = DefaultWardenAgentConfiguration
	}

	err = config.validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating config")
	}

	return config, nil
}

func (c Config) validate() error {
	if c.AssetsDir == "" {
		return bosherr.New("Must provide non-empty assets_dir")
	}

	if c.ReposDir == "" {
		return bosherr.New("Must provide non-empty repos_dir")
	}

	if c.Blobstore.Type != bpprov.BlobstoreConfigTypeLocal {
		return bosherr.New("Blobstore type must be local")
	}

	err := c.Blobstore.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating blobstore configuration")
	}

	return nil
}
