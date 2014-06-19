package manifest

import (
	bosherr "bosh/errors"

	bputil "boshprovisioner/util"
)

// SyntaxValidator parses and saves all manifest values to determine
// their syntactic validity. Determining if individual values make sense
// in a greater context (within a deployment or a job) is outside of scope.
// e.g. - can watch time string value be parsed into a time range?
type SyntaxValidator struct {
	deployment *Deployment
}

func NewSyntaxValidator(manifest *Manifest) SyntaxValidator {
	if manifest == nil {
		panic("Expected manifest to not be nil")
	}

	return SyntaxValidator{deployment: &manifest.Deployment}
}

func (v SyntaxValidator) Validate() error {
	if v.deployment.Name == "" {
		return bosherr.New("Missing deployment name")
	}

	err := v.validateUpdate(&v.deployment.Update)
	if err != nil {
		return bosherr.WrapError(err, "Deployment update")
	}

	for i, net := range v.deployment.Networks {
		err := v.validateNetwork(&v.deployment.Networks[i])
		if err != nil {
			return bosherr.WrapError(err, "Network %s (%d)", net.Name, i)
		}
	}

	for i, release := range v.deployment.Releases {
		err := v.validateRelease(&v.deployment.Releases[i])
		if err != nil {
			return bosherr.WrapError(err, "Release %s (%d)", release.Name, i)
		}
	}

	err = v.validateCompilation(&v.deployment.Compilation)
	if err != nil {
		return bosherr.WrapError(err, "Compilation")
	}

	for i, job := range v.deployment.Jobs {
		err := v.validateJob(&v.deployment.Jobs[i])
		if err != nil {
			return bosherr.WrapError(err, "Job %s (%d)", job.Name, i)
		}
	}

	props, err := bputil.NewStringKeyed().ConvertMap(v.deployment.PropertiesRaw)
	if err != nil {
		return bosherr.WrapError(err, "Deployment properties")
	}

	v.deployment.Properties = props

	return nil
}

func (v SyntaxValidator) validateNetwork(network *Network) error {
	if network.Name == "" {
		return bosherr.New("Missing network name")
	}

	return v.validateNetworkType(network.Type)
}

func (v SyntaxValidator) validateNetworkType(networkType string) error {
	if networkType == "" {
		return bosherr.New("Missing network type")
	}

	for _, t := range NetworkTypes {
		if networkType == t {
			return nil
		}
	}

	return bosherr.New("Unknown network type %s", networkType)
}

func (v SyntaxValidator) validateRelease(release *Release) error {
	if release.Name == "" {
		return bosherr.New("Missing release name")
	}

	if release.Version == "" {
		return bosherr.New("Missing release version")
	}

	if release.URL == "" {
		return bosherr.New("Missing release URL")
	}

	return nil
}

func (v SyntaxValidator) validateCompilation(compilation *Compilation) error {
	if compilation.NetworkName == "" {
		return bosherr.New("Missing network name")
	}

	return nil
}

func (v SyntaxValidator) validateJob(job *Job) error {
	if job.Name == "" {
		return bosherr.New("Missing job name")
	}

	if job.Template != nil {
		return bosherr.New("'template' is deprecated in favor of 'templates'")
	}

	err := v.validateUpdate(&job.Update)
	if err != nil {
		return bosherr.WrapError(err, "Update")
	}

	props, err := bputil.NewStringKeyed().ConvertMap(job.PropertiesRaw)
	if err != nil {
		return bosherr.WrapError(err, "Properties")
	}

	job.Properties = props

	for i, na := range job.NetworkAssociations {
		err := v.validateNetworkAssociation(&job.NetworkAssociations[i])
		if err != nil {
			return bosherr.WrapError(err, "Network association %s (%d)", na.NetworkName, i)
		}
	}

	return nil
}

// validateUpdate validates deployment level or job level update section
func (v SyntaxValidator) validateUpdate(update *Update) error {
	if update.CanaryWatchTimeRaw != nil {
		watchTime, err := NewWatchTimeFromString(*update.CanaryWatchTimeRaw)
		if err != nil {
			return bosherr.WrapError(err, "Canary watch time")
		}

		update.CanaryWatchTime = &watchTime
	}

	if update.UpdateWatchTimeRaw != nil {
		watchTime, err := NewWatchTimeFromString(*update.UpdateWatchTimeRaw)
		if err != nil {
			return bosherr.WrapError(err, "Update watch time")
		}

		update.UpdateWatchTime = &watchTime
	}

	return nil
}

func (v SyntaxValidator) validateNetworkAssociation(na *NetworkAssociation) error {
	if na.NetworkName == "" {
		return bosherr.New("Missing network name")
	}

	if na.StaticIPsRaw != nil {
		ips, err := NewIPsFromStrings(na.StaticIPsRaw)
		if err != nil {
			return bosherr.WrapError(err, "Static IPs")
		}

		na.StaticIPs = ips
	}

	return nil
}
