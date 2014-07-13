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

func (r CTTJRepository) FindByTemplate(template bpdep.Template) (ReleaseJobRecord, bool, error) {
	var rec ReleaseJobRecord

	err := r.index.Find(r.templateKey(template), &rec)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return rec, false, nil
		}

		return rec, false, bosherr.WrapError(err, "Finding dep-template -> release-job record")
	}

	return rec, true, nil
}

func (r CTTJRepository) SaveForJob(release bprel.Release, job bprel.Job) (ReleaseJobRecord, error) {
	// todo redundant info stored in value
	rec := ReleaseJobRecord{
		ReleaseName:    release.Name,
		ReleaseVersion: release.Version,

		JobName:        job.Name,
		JobVersion:     job.Version,
		JobFingerprint: job.Fingerprint,
	}

	err := r.index.Save(r.jobKey(release, job), rec)
	if err != nil {
		return rec, bosherr.WrapError(err, "Saving dep-template -> release-job record")
	}

	return rec, nil
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
