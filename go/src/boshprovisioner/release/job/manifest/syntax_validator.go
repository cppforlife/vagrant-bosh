package manifest

import (
	bosherr "bosh/errors"

	bputil "boshprovisioner/util"
)

type SyntaxValidator struct {
	job *Job
}

func NewSyntaxValidator(manifest *Manifest) SyntaxValidator {
	if manifest == nil {
		panic("Expected manifest to not be nil")
	}

	return SyntaxValidator{job: &manifest.Job}
}

func (v SyntaxValidator) Validate() error {
	for name, propDef := range v.job.PropertyMappings {
		propDef, err := v.validatePropDef(propDef)
		if err != nil {
			return bosherr.WrapError(err, "Property %s", name)
		}

		v.job.PropertyMappings[name] = propDef
	}

	return nil
}

func (v SyntaxValidator) validatePropDef(propDef PropertyDefinition) (PropertyDefinition, error) {
	def, err := bputil.NewStringKeyed().ConvertInterface(propDef.DefaultRaw)
	if err != nil {
		return propDef, bosherr.WrapError(err, "Default")
	}

	propDef.Default = def

	return propDef, nil
}
