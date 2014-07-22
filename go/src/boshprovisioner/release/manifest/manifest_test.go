package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "boshprovisioner/release/manifest"
)

var _ = Describe("Manifest", func() {
	Describe("NewManifestFromBytes", func() {
		It("returns manifest with decoded version/fingerprint/sha1 for jobs and packages", func() {
			manifestBytes := []byte(`
name: bosh
version: 77

commit_hash: bbe5476c
uncommitted_changes: true

packages:
- name: registry
  version: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  fingerprint: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  sha1: !binary |-
    NmVhYTZjOTYxZWFjN2JkOTk0ZDE2NDRhZDQwNWIzMzk1NDIwZWNhZg==
  dependencies: []

jobs:
- name: powerdns
  version: !binary |-
    MGI4MGIzYzE5OGJmN2FiYzZjODEyNjIwMTNkZTQ5NDM2OWZkMjViNg==
  fingerprint: !binary |-
    MGI4MGIzYzE5OGJmN2FiYzZjODEyNjIwMTNkZTQ5NDM2OWZkMjViNg==
  sha1: !binary |-
    YWI5NzA5YmVhYjViZTBmYjYyYTJkMWYzYzg4ZDA2YzliNGJkZWM2NQ==
`)

			manifest, err := NewManifestFromBytes(manifestBytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(manifest.Release.Packages[0]).To(Equal(Package{
				Name: "registry",

				VersionRaw: "ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==",
				Version:    "dd1ba330bc44b3181b263383b8e4252d7051deca",

				FingerprintRaw: "ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==",
				Fingerprint:    "dd1ba330bc44b3181b263383b8e4252d7051deca",

				SHA1Raw: "NmVhYTZjOTYxZWFjN2JkOTk0ZDE2NDRhZDQwNWIzMzk1NDIwZWNhZg==",
				SHA1:    "6eaa6c961eac7bd994d1644ad405b3395420ecaf",

				DependencyNames: []DependencyName{},
			}))

			Expect(manifest.Release.Jobs[0]).To(Equal(Job{
				Name: "powerdns",

				VersionRaw: "MGI4MGIzYzE5OGJmN2FiYzZjODEyNjIwMTNkZTQ5NDM2OWZkMjViNg==",
				Version:    "0b80b3c198bf7abc6c81262013de494369fd25b6",

				FingerprintRaw: "MGI4MGIzYzE5OGJmN2FiYzZjODEyNjIwMTNkZTQ5NDM2OWZkMjViNg==",
				Fingerprint:    "0b80b3c198bf7abc6c81262013de494369fd25b6",

				SHA1Raw: "YWI5NzA5YmVhYjViZTBmYjYyYTJkMWYzYzg4ZDA2YzliNGJkZWM2NQ==",
				SHA1:    "ab9709beab5be0fb62a2d1f3c88d06c9b4bdec65",
			}))
		})

		It("returns manifest with given version/fingerprint/sha1 for jobs and packages if they were not base64 binary encoded", func() {
			manifestBytes := []byte(`
name: bosh
version: 77

commit_hash: bbe5476c
uncommitted_changes: true

packages:
- name: registry
  version: dd1ba330bc44b3181b263383b8e4252d7051deca
  fingerprint: dd1ba330bc44b3181b263383b8e4252d7051deca
  sha1: 6eaa6c961eac7bd994d1644ad405b3395420ecaf
  dependencies: []

jobs:
- name: powerdns
  version: 0b80b3c198bf7abc6c81262013de494369fd25b6
  fingerprint: 0b80b3c198bf7abc6c81262013de494369fd25b6
  sha1: ab9709beab5be0fb62a2d1f3c88d06c9b4bdec65
`)

			manifest, err := NewManifestFromBytes(manifestBytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(manifest.Release.Packages[0]).To(Equal(Package{
				Name: "registry",

				VersionRaw: "dd1ba330bc44b3181b263383b8e4252d7051deca",
				Version:    "dd1ba330bc44b3181b263383b8e4252d7051deca",

				FingerprintRaw: "dd1ba330bc44b3181b263383b8e4252d7051deca",
				Fingerprint:    "dd1ba330bc44b3181b263383b8e4252d7051deca",

				SHA1Raw: "6eaa6c961eac7bd994d1644ad405b3395420ecaf",
				SHA1:    "6eaa6c961eac7bd994d1644ad405b3395420ecaf",

				DependencyNames: []DependencyName{},
			}))

			Expect(manifest.Release.Jobs[0]).To(Equal(Job{
				Name: "powerdns",

				VersionRaw: "0b80b3c198bf7abc6c81262013de494369fd25b6",
				Version:    "0b80b3c198bf7abc6c81262013de494369fd25b6",

				FingerprintRaw: "0b80b3c198bf7abc6c81262013de494369fd25b6",
				Fingerprint:    "0b80b3c198bf7abc6c81262013de494369fd25b6",

				SHA1Raw: "ab9709beab5be0fb62a2d1f3c88d06c9b4bdec65",
				SHA1:    "ab9709beab5be0fb62a2d1f3c88d06c9b4bdec65",
			}))
		})
	})
})
