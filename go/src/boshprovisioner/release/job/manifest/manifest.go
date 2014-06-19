// Package manifest represents internal structure of a release job.
package manifest

import (
	bosherr "bosh/errors"
	boshsys "bosh/system"
	"github.com/fraenkel/candiedyaml"
)

type Manifest struct {
	Job Job
}

type Job struct {
	Name string `yaml:"name"`

	TemplateNames TemplateNames `yaml:"templates"`

	PackageNames []string `yaml:"packages"`

	PropertyMappings PropertyMappings `yaml:"properties"`
}

type TemplateNames map[string]string

type PropertyMappings map[string]PropertyDefinition

type PropertyDefinition struct {
	Description string `yaml:"description"`

	// Non-raw field is populated by the validator.
	DefaultRaw interface{} `yaml:"default"`
	Default    interface{}
}

func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	bytes, err := fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Reading manifest %s", path)
	}

	return NewManifestFromBytes(bytes)
}

func NewManifestFromBytes(bytes []byte) (Manifest, error) {
	var manifest Manifest
	var job Job

	err := candiedyaml.Unmarshal(bytes, &job)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Parsing job")
	}

	manifest.Job = job

	err = NewSyntaxValidator(&manifest).Validate()
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Validating manifest syntactically")
	}

	return manifest, nil
}

/*
# Example for job.MF
name: dummy

templates:
  dummy_ctl: bin/dummy_ctl

packages:
- dummy_package
- dummy_package2

properties:
  dummy_value:
    description: Some value for the dummy job
    default: 300
*/
