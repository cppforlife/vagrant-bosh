package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "boshprovisioner/release/job/manifest"
)

var _ = Describe("Manifest", func() {
	Describe("NewManifestFromBytes", func() {
		It("returns manifest with property definiton default that have string keys", func() {
			manifestBytes := []byte(`
properties:
  key:
    default:
      prop:
        nest-prop: instance-val
      props:
      - name: nest-prop
`)

			manifest, err := NewManifestFromBytes(manifestBytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(manifest.Job.PropertyMappings).To(HaveLen(1))

			for _, propDef := range manifest.Job.PropertyMappings {
				// candiedyaml unmarshals manifest to map[interface{}]interface{}
				// (encoding/json unmarshals manifest to map[string]interface{})
				Expect(propDef.Default).To(Equal(map[string]interface{}{
					"prop": map[string]interface{}{
						"nest-prop": "instance-val",
					},
					"props": []interface{}{
						map[string]interface{}{"name": "nest-prop"},
					},
				}))
			}
		})
	})
})
