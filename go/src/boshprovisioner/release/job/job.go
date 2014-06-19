package job

import (
	bpreljobman "boshprovisioner/release/job/manifest"
)

type Job struct {
	Manifest bpreljobman.Manifest

	Name string

	MonitTemplate Template

	Templates []Template

	// Runtime package dependencies for this job
	Packages []Package

	Properties []Property
}

type Template struct {
	SrcPathEnd string
	DstPathEnd string // End of the path on the VM

	Path string
}

type Package struct {
	Name string
}

type Property struct {
	Name        string
	Description string

	Default interface{}
}

// populateFromManifest populates job information interpreted from job manifest.
func (j *Job) populateFromManifest(manifest bpreljobman.Manifest) {
	j.populateJob(manifest.Job)
	j.populateTemplates(manifest.Job.TemplateNames)
	j.populatePackages(manifest.Job.PackageNames)
	j.populateProperties(manifest.Job.PropertyMappings)
	j.Manifest = manifest
}

func (j *Job) populateJob(manJob bpreljobman.Job) {
	j.Name = manJob.Name
}

func (j *Job) populateTemplates(manTemplateNames bpreljobman.TemplateNames) {
	j.MonitTemplate = Template{
		SrcPathEnd: "monit",
		DstPathEnd: "monit",
	}

	for srcPathEnd, dstPathEnd := range manTemplateNames {
		template := Template{
			SrcPathEnd: srcPathEnd,
			DstPathEnd: dstPathEnd,
		}

		j.Templates = append(j.Templates, template)
	}
}

func (j *Job) populatePackages(manPackageNames []string) {
	for _, name := range manPackageNames {
		j.Packages = append(j.Packages, Package{Name: name})
	}
}

func (j *Job) populateProperties(manPropMappings bpreljobman.PropertyMappings) {
	for propName, propDef := range manPropMappings {
		property := Property{
			Name:        propName,
			Description: propDef.Description,

			Default: propDef.Default,
		}

		j.Properties = append(j.Properties, property)
	}
}
