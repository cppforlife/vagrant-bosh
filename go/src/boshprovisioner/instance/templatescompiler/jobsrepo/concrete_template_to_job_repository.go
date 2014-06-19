package jobsrepo

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpindex "boshprovisioner/index"
	bprel "boshprovisioner/release"
)

type CTTJRepository struct {
	index  bpindex.Index
	logger boshlog.Logger
}

type templateToJobKey struct {
	JobName        string
	ReleaseName    string
	ReleaseVersion string
}

func NewConcreteTemplateToJobRepository(
	index bpindex.Index,
	logger boshlog.Logger,
) CTTJRepository {
	return CTTJRepository{
		index:  index,
		logger: logger,
	}
}

func (r CTTJRepository) FindByTemplate(template bpdep.Template) (bprel.Job, bool, error) {
	var job bprel.Job

	err := r.index.Find(r.templateKey(template), &job)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return job, false, nil
		}

		return job, false, bosherr.WrapError(err, "Finding dep-template -> rel-job record")
	}

	return job, true, nil
}

func (r CTTJRepository) SaveForJob(release bprel.Release, job bprel.Job) error {
	err := r.index.Save(r.jobKey(release, job), job)
	if err != nil {
		return bosherr.WrapError(err, "Saving dep-template -> rel-job record")
	}

	return nil
}

func (r CTTJRepository) templateKey(template bpdep.Template) templateToJobKey {
	if template.Release == nil {
		panic("Expected template.Release to not be nil")
	}

	return templateToJobKey{
		JobName:        template.Name,
		ReleaseName:    template.Release.Name,
		ReleaseVersion: template.Release.Version,
	}
}

// todo should job point back to release
func (r CTTJRepository) jobKey(release bprel.Release, job bprel.Job) templateToJobKey {
	return templateToJobKey{
		JobName:        job.Name,
		ReleaseName:    release.Name,
		ReleaseVersion: release.Version,
	}
}
