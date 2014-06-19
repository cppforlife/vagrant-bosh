package erbrenderer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bpdep "boshprovisioner/deployment"
	. "boshprovisioner/instance/templatescompiler/erbrenderer"
	bpreljob "boshprovisioner/release/job"
)

var _ = Describe("RenderProperties", func() {
	var (
		job      bpreljob.Job
		instance bpdep.Instance
		props    RenderProperties
	)

	JustBeforeEach(func() {
		props = NewRenderProperties(job, instance)
	})

	Describe("AsMap", func() {
		Context("when job specifies a default for a nested property", func() {
			BeforeEach(func() {
				job = bpreljob.Job{
					Properties: []bpreljob.Property{
						bpreljob.Property{
							Name:    "prop.nest-prop",
							Default: "job-val",
						},
					},
				}
			})

			Context("when instance specifies nested property", func() {
				BeforeEach(func() {
					instance = bpdep.Instance{
						Properties: map[string]interface{}{
							"prop": map[string]interface{}{
								"nest-prop": "instance-val",
							},
						},
					}
				})

				It("returns map with property value from job", func() {
					Expect(props.AsMap()).To(Equal(map[string]interface{}{
						"prop": map[string]interface{}{
							"nest-prop": "instance-val",
						},
					}))
				})
			})

			Context("when instance does not specify nested property", func() {
				Context("when first nesting level is specified", func() {
					BeforeEach(func() {
						instance = bpdep.Instance{
							Properties: map[string]interface{}{
								"prop": map[string]interface{}{},
							},
						}
					})

					It("returns map with default property from job", func() {
						Expect(props.AsMap()).To(Equal(map[string]interface{}{
							"prop": map[string]interface{}{
								"nest-prop": "job-val",
							},
						}))
					})
				})

				Context("when first nesting level is not specified", func() {
					BeforeEach(func() {
						instance = bpdep.Instance{}
					})

					It("returns map with default property from job", func() {
						Expect(props.AsMap()).To(Equal(map[string]interface{}{
							"prop": map[string]interface{}{
								"nest-prop": "job-val",
							},
						}))
					})
				})
			})
		})

		Context("when job does not specify a default (nil) for a nested property", func() {
			BeforeEach(func() {
				job = bpreljob.Job{
					Properties: []bpreljob.Property{
						bpreljob.Property{
							Name:    "prop.nest-prop",
							Default: nil,
						},
					},
				}
			})

			Context("when instance specifies nested property", func() {
				BeforeEach(func() {
					instance = bpdep.Instance{
						Properties: map[string]interface{}{
							"prop": map[string]interface{}{
								"nest-prop": "instance-val",
							},
						},
					}
				})

				It("returns map with property value from instance", func() {
					Expect(props.AsMap()).To(Equal(map[string]interface{}{
						"prop": map[string]interface{}{
							"nest-prop": "instance-val",
						},
					}))
				})
			})

			Context("when instance does not specify nested property", func() {
				BeforeEach(func() {
					instance = bpdep.Instance{
						Properties: map[string]interface{}{
							"prop": map[string]interface{}{},
						},
					}
				})

				It("returns map with default property from job", func() {
					Expect(props.AsMap()).To(Equal(map[string]interface{}{
						"prop": map[string]interface{}{
							"nest-prop": nil,
						},
					}))
				})
			})
		}) //~
	})
})
