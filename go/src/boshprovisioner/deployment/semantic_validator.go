package deployment

import (
	bosherr "bosh/errors"
)

// SemanticValidator validates deployment to determine if it represents a meaningful state.
// e.g. - is each job template associated with a release?
//      - are there enough static ips for each job instance?
type SemanticValidator struct {
	deployment Deployment
}

func NewSemanticValidator(deployment Deployment) SemanticValidator {
	return SemanticValidator{deployment: deployment}
}

func (v SemanticValidator) Validate() error {
	for _, net := range v.deployment.Networks {
		err := v.validateNetwork(net)
		if err != nil {
			return bosherr.WrapError(err, "Network %s", net.Name)
		}
	}

	for _, release := range v.deployment.Releases {
		err := v.validateRelease(release)
		if err != nil {
			return bosherr.WrapError(err, "Release %s", release.Name)
		}
	}

	err := v.validateInstance(v.deployment.CompilationInstance)
	if err != nil {
		return bosherr.WrapError(err, "Compilation instance")
	}

	for _, job := range v.deployment.Jobs {
		err := v.validateJob(job)
		if err != nil {
			return bosherr.WrapError(err, "Job %s", job.Name)
		}
	}

	return nil
}

func (v SemanticValidator) validateNetwork(network Network) error {
	if network.Type == NetworkTypeManual {
		return bosherr.New("Manual networking is not supported")
	}

	return nil
}

func (v SemanticValidator) validateRelease(release Release) error {
	if release.Version == "latest" {
		return bosherr.New("Version 'latest' is not supported")
	}

	return nil
}

func (v SemanticValidator) validateJob(job Job) error {
	for _, template := range job.Templates {
		err := v.validateTemplate(template)
		if err != nil {
			return bosherr.WrapError(err, "Template %s", template.Name)
		}
	}

	for _, instance := range job.Instances {
		err := v.validateInstance(instance)
		if err != nil {
			return bosherr.WrapError(err, "Instance %d", instance.Index)
		}
	}

	return nil
}

func (v SemanticValidator) validateTemplate(template Template) error {
	if template.Release == nil {
		return bosherr.New("Missing associated release")
	}

	return nil
}

func (v SemanticValidator) validateInstance(instance Instance) error {
	for i, na := range instance.NetworkAssociations {
		err := v.validateNetworkAssociation(na)
		if err != nil {
			return bosherr.WrapError(err, "Network association %d", i)
		}
	}

	return nil
}

func (v SemanticValidator) validateNetworkAssociation(na NetworkAssociation) error {
	if na.Network == nil {
		return bosherr.New("Missing associated network")
	}

	if na.MustHaveStaticIP && na.StaticIP == nil {
		return bosherr.New("Missing static IP assignment")
	}

	return nil
}
