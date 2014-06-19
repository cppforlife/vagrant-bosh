// Package manifest represents internal structure of a release.
package manifest

import (
	bosherr "bosh/errors"
	boshsys "bosh/system"
	"github.com/fraenkel/candiedyaml"
)

type Manifest struct {
	Release Release
}

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	Jobs     []Job     `yaml:"jobs"`
	Packages []Package `yaml:"packages"`

	CommitHash         string `yaml:"commit_hash"`
	UncommittedChanges bool   `yaml:"uncommitted_changes"`
}

type Job struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`

	DependencyNames []DependencyName `yaml:"dependencies"`
}

type Package struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`

	DependencyNames []DependencyName `yaml:"dependencies"`
}

func (p Package) DependencyName() DependencyName {
	return DependencyName(p.Name)
}

type DependencyName string

func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	bytes, err := fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Reading manifest %s", path)
	}

	var release Release

	err = candiedyaml.Unmarshal(bytes, &release)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Parsing release")
	}

	return Manifest{Release: release}, nil
}

/*
# Example for release.MF
name: bosh
version: 77

commit_hash: bbe5476c
uncommitted_changes: true

packages:
- name: registry
  version: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  sha1: !binary |-
    NmVhYTZjOTYxZWFjN2JkOTk0ZDE2NDRhZDQwNWIzMzk1NDIwZWNhZg==
  fingerprint: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  dependencies:
  - libpq
  - mysql
  - ruby

jobs:
- name: powerdns
  version: !binary |-
    MGI4MGIzYzE5OGJmN2FiYzZjODEyNjIwMTNkZTQ5NDM2OWZkMjViNg==
  fingerprint: !binary |-
    MGI4MGIzYzE5OGJmN2FiYzZjODEyNjIwMTNkZTQ5NDM2OWZkMjViNg==
  sha1: !binary |-
    YWI5NzA5YmVhYjViZTBmYjYyYTJkMWYzYzg4ZDA2YzliNGJkZWM2NQ==
*/
