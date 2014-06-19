package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "boshprovisioner/deployment/manifest"
)

var _ = Describe("Manifest", func() {
	Describe("NewManifestFromBytes", func() {
		It("returns manifest with deployment properties that have string keys", func() {
			manifestBytes := []byte(`
name: fake-deployment

networks:
- name: net1
  type: dynamic

compilation:
  network: net1

properties:
  prop:
    nest-prop: instance-val
  props:
  - name: nest-prop
`)

			manifest, err := NewManifestFromBytes(manifestBytes)
			Expect(err).ToNot(HaveOccurred())

			depProps := manifest.Deployment.Properties

			// candiedyaml unmarshals manifest to map[interface{}]interface{}
			// (encoding/json unmarshals manifest to map[string]interface{})
			Expect(depProps).To(Equal(Properties(
				map[string]interface{}{
					"prop": map[string]interface{}{
						"nest-prop": "instance-val",
					},
					"props": []interface{}{
						map[string]interface{}{"name": "nest-prop"},
					},
				},
			)))
		})

		It("returns manifest with job properties that have string keys", func() {
			manifestBytes := []byte(`
name: fake-deployment

networks:
- name: net1
  type: dynamic

compilation:
  network: net1

jobs:
- name: job-1
  properties:
    prop:
      nest-prop: instance-val
    props:
    - name: nest-prop
`)

			manifest, err := NewManifestFromBytes(manifestBytes)
			Expect(err).ToNot(HaveOccurred())

			jobProps := manifest.Deployment.Jobs[0].Properties

			Expect(jobProps).To(Equal(Properties(
				map[string]interface{}{
					"prop": map[string]interface{}{
						"nest-prop": "instance-val",
					},
					"props": []interface{}{
						map[string]interface{}{"name": "nest-prop"},
					},
				},
			)))
		})
	})
})
