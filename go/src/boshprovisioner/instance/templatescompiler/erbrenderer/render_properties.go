package erbrenderer

import (
	"encoding/json"
	"strings"

	bosherr "bosh/errors"

	bpdep "boshprovisioner/deployment"
	bpreljob "boshprovisioner/release/job"
)

type RenderProperties struct {
	relJob   bpreljob.Job
	instance bpdep.Instance
}

func NewRenderProperties(relJob bpreljob.Job, instance bpdep.Instance) RenderProperties {
	return RenderProperties{relJob: relJob, instance: instance}
}

// AsMap returns job and instance properties merged together.
func (p RenderProperties) AsMap() (map[string]interface{}, error) {
	result, err := p.deepCopyInstanceProperties()
	if err != nil {
		return result, err
	}

	for _, prop := range p.relJob.Properties {
		p.copyProperty(prop.Name, prop.Default, result)
	}

	return result, nil
}

// copyProperty fills in property in dst with value if it's not already set.
func (p RenderProperties) copyProperty(path string, value interface{}, dst map[string]interface{}) {
	pathParts := strings.Split(path, ".")

	for i, part := range pathParts {
		if dstNested, ok := dst[part]; ok { // Found section; check if modified
			if dstNestedMap, ok := dstNested.(map[string]interface{}); ok {
				dst = dstNestedMap
			} else {
				break // Not a section; cannot modify
			}
		} else if len(pathParts)-1 == i { // Last property path part
			dst[part] = value
		} else {
			m := map[string]interface{}{}
			dst[part] = m
			dst = m
		}
	}
}

// deepCopyInstanceProperties makes a deep copy of instance properties.
// Always returns an initialized map even if instance properties are nil.
func (p RenderProperties) deepCopyInstanceProperties() (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if p.instance.Properties == nil {
		return result, nil
	}

	bytes, err := json.Marshal(p.instance.Properties)
	if err != nil {
		return result, bosherr.WrapError(err, "Marshalling instance properties")
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, bosherr.WrapError(err, "Unmarshalling instance properties")
	}

	return result, nil
}
