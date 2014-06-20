package manifest

import (
	"encoding/json"

	bosherr "bosh/errors"
)

func (d Deployment) InstanceWatchTime(job Job, i int) WatchTime {
	var canaries int

	if job.Update.Canaries != nil {
		canaries = *job.Update.Canaries
	} else if d.Update.Canaries != nil {
		canaries = *d.Update.Canaries
	}

	if canaries > i {
		return d.CanaryWatchTime(job)
	}

	return d.UpdateWatchTime(job)
}

func (d Deployment) CanaryWatchTime(job Job) WatchTime {
	if job.Update.CanaryWatchTime != nil {
		return *job.Update.CanaryWatchTime
	} else if d.Update.CanaryWatchTime != nil {
		return *d.Update.CanaryWatchTime
	}

	return DefaultWatchTime
}

func (d Deployment) UpdateWatchTime(job Job) WatchTime {
	if job.Update.UpdateWatchTime != nil {
		return *job.Update.UpdateWatchTime
	} else if d.Update.UpdateWatchTime != nil {
		return *d.Update.UpdateWatchTime
	}

	return DefaultWatchTime
}

func (d Deployment) InstanceProperties(job Job, i int) Properties {
	result, err := job.deepCopyProperties()
	if err != nil {
		panic("Deep copying job properties")
	}

	for name, value := range d.Properties {
		if _, ok := result[name]; !ok {
			result[name] = value
		}
	}

	return result
}

// deepCopyJobProperties makes a deep copy of job properties.
// Always returns an initialized map even if job properties are nil.
func (j Job) deepCopyProperties() (Properties, error) {
	result := map[string]interface{}{}

	if j.Properties == nil {
		return result, nil
	}

	bytes, err := json.Marshal(j.Properties)
	if err != nil {
		return result, bosherr.WrapError(err, "Marshalling job properties")
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, bosherr.WrapError(err, "Unmarshalling job properties")
	}

	return result, nil
}
