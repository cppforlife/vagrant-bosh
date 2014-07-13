package jobsrepo

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpindex "boshprovisioner/index"
	bprel "boshprovisioner/release"
)

type CJRepository struct {
	index  bpindex.Index
	logger boshlog.Logger
}

type jobToJobKey struct {
	JobName    string
	JobVersion string

	// Fingerprint of a job captures its dependenices; however,
	// that does not capture package changes in an associated release.
	JobFingerprint string
}

func NewConcreteJobsRepository(
	index bpindex.Index,
	logger boshlog.Logger,
) CJRepository {
	return CJRepository{
		index:  index,
		logger: logger,
	}
}

func (r CJRepository) Find(job bprel.Job) (JobRecord, bool, error) {
	var record JobRecord

	err := r.index.Find(r.key(job), &record)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return record, false, nil
		}

		return record, false, bosherr.WrapError(err, "Finding job record")
	}

	return record, true, nil
}

func (r CJRepository) FindByReleaseJob(rec ReleaseJobRecord) (JobRecord, bool, error) {
	var record JobRecord

	key := jobToJobKey{
		JobName:        rec.JobName,
		JobVersion:     rec.JobVersion,
		JobFingerprint: rec.JobFingerprint,
	}

	err := r.index.Find(key, &record)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return record, false, nil
		}

		return record, false, bosherr.WrapError(err, "Finding job record")
	}

	return record, true, nil
}

func (r CJRepository) Save(job bprel.Job, record JobRecord) error {
	err := r.index.Save(r.key(job), record)
	if err != nil {
		return bosherr.WrapError(err, "Saving job record")
	}

	return nil
}

func (r CJRepository) key(job bprel.Job) jobToJobKey {
	return jobToJobKey{
		JobName:        job.Name,
		JobVersion:     job.Version,
		JobFingerprint: job.Fingerprint,
	}
}
