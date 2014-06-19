package main

import (
	"encoding/json"

	bosherr "bosh/errors"
	boshsys "bosh/system"

	bpprov "boshprovisioner/provisioner"
)

type Config struct {
	ManifestPath string `json:"manifest_path"`

	// Assets dir is used as a temporary location to transfer files from host to guest.
	// It will not be created since assets already must be present.
	AssetsDir string `json:"assets_dir"`

	// Repos dir is mainly used to record what was placed in the blobstore.
	// It will be created if it does not exist.
	ReposDir string `json:"repos_dir"`

	// e.g. "https://user:password@127.0.0.1:4321/agent"
	Mbus string `json:"mbus"`

	Blobstore bpprov.BlobstoreConfig `json:"blobstore"`
}

func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapError(err, "Reading config %s", path)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	err = config.validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating config")
	}

	return config, nil
}

func (c Config) validate() error {
	if c.ManifestPath == "" {
		return bosherr.New("Must provide non-empty manifest_path")
	}

	if c.AssetsDir == "" {
		return bosherr.New("Must provide non-empty assets_dir")
	}

	if c.ReposDir == "" {
		return bosherr.New("Must provide non-empty repos_dir")
	}

	if c.Mbus == "" {
		return bosherr.New("Must provide non-empty mbus")
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
