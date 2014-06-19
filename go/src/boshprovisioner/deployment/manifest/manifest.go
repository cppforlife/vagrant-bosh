// Package manifest represents structure
// of a user entered YAML deployment manifest.
package manifest

import (
	gonet "net"

	bosherr "bosh/errors"
	boshsys "bosh/system"
	"github.com/fraenkel/candiedyaml"
)

type Manifest struct {
	Deployment Deployment
}

type Deployment struct {
	Name string `yaml:"name"`

	Releases []Release `yaml:"releases"`

	Networks []Network `yaml:"networks"`

	Compilation Compilation `yaml:"compilation"`

	// Deployment-wide update config can be
	// overwritten by job-specific update config
	Update Update `yaml:"update"`

	Jobs []Job `yaml:"jobs"`

	// Global properties.
	// Non-raw field is populated by the validator.
	PropertiesRaw map[interface{}]interface{} `yaml:"properties"`
	Properties    Properties
}

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	// Not offical BOSH manifest construct
	URL string `yaml:"url"`
}

const (
	NetworkTypeManual  = "manual"
	NetworkTypeDynamic = "dynamic"
	NetworkTypeVip     = "vip"
)

var NetworkTypes = []string{NetworkTypeManual, NetworkTypeDynamic, NetworkTypeVip}

type Network struct {
	Name string `yaml:"name"`

	// e.g. manual, dynamic, vip
	Type string `yaml:"type"`
}

type Compilation struct {
	NetworkName string `yaml:"network"`
}

type Job struct {
	Name      string `yaml:"name"`
	Instances int    `yaml:"instances"`

	Update Update `yaml:"update"`

	// Deprecated in favor of Templates
	Template interface{} `yaml:"template"`

	Templates []Template `yaml:"templates"`

	// Job specific properties that override global properties.
	// Non-raw field is populated by the validator.
	PropertiesRaw map[interface{}]interface{} `yaml:"properties"`
	Properties    Properties

	NetworkAssociations []NetworkAssociation `yaml:"networks"`
}

type Template struct {
	Name        string `yaml:"name"`
	ReleaseName string `yaml:"release"`
}

type Properties map[string]interface{}

type Update struct {
	// Integer pointers to determine absence
	Canaries    *int `yaml:"canaries"`
	MaxInFlight *int `yaml:"max_in_flight"`

	// String pointers to determine absence
	CanaryWatchTimeRaw *string `yaml:"canary_watch_time"`
	UpdateWatchTimeRaw *string `yaml:"update_watch_time"`

	// Populated by the validator
	CanaryWatchTime *WatchTime
	UpdateWatchTime *WatchTime
}

type NetworkAssociation struct {
	NetworkName string `yaml:"name"`

	// Non-raw field populated by the validator.
	StaticIPsRaw []string `yaml:"static_ips"`
	StaticIPs    []gonet.IP
}

// NewManifestFromPath returns manifest read from the file system.
// Before returning manifest is syntactically validated.
func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	bytes, err := fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Reading manifest %s", path)
	}

	return NewManifestFromBytes(bytes)
}

// NewManifestFromBytes returns manifest built from given bytes.
// Before returning manifest is syntactically validated.
func NewManifestFromBytes(bytes []byte) (Manifest, error) {
	var manifest Manifest
	var deployment Deployment

	err := candiedyaml.Unmarshal(bytes, &deployment)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Parsing deployment")
	}

	manifest.Deployment = deployment

	err = NewSyntaxValidator(&manifest).Validate()
	if err != nil {
		return manifest, bosherr.WrapError(err, "Validating manifest syntactically")
	}

	return manifest, nil
}
