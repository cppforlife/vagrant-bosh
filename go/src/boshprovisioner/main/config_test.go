package main_test

import (
	fakesys "bosh/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "boshprovisioner/main"
	bpvm "boshprovisioner/vm"
)

var _ = Describe("NewConfigFromPath", func() {
	var (
		fs *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
	})

	It("defautls null values to agent provisioner config defaults", func() {
		configJSON := `{
      "manifest_path": "fake-manifest-path",
      "assets_dir": "fake-assets-dir",
      "repos_dir": "fake-repos-dir",
      "blobstore": {
        "provider": "local",
        "options": {
          "blobstore_path": "fake-blobstore-path"
        }
      },
      "vm_provisioner": {
        "agent_provisioner": {
          "infrastructure": null,
          "platform": null,
          "configuration": null,
          "mbus": null
        }
      }
    }`

		err := fs.WriteFileString("/tmp/config", configJSON)
		Expect(err).ToNot(HaveOccurred())

		config, err := NewConfigFromPath("/tmp/config", fs)
		Expect(err).ToNot(HaveOccurred())

		Expect(config.VMProvisioner.AgentProvisioner).To(Equal(
			bpvm.AgentProvisionerConfig{
				Infrastructure: "warden",
				Platform:       "ubuntu",

				Configuration: map[string]interface{}{
					"Platform": map[string]interface{}{
						"Linux": map[string]interface{}{
							"UseDefaultTmpDir":              true,
							"UsePreformattedPersistentDisk": true,
							"BindMountPersistentDisk":       true,
						},
					},
				},

				Mbus: "https://user:password@127.0.0.1:4321/agent",
			},
		))
	})
})
