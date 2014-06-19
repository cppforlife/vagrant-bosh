package jobsrepo

import (
	bpdep "boshprovisioner/deployment"
	bprel "boshprovisioner/release"
)

type JobRecord struct {
	BlobID string
	SHA1   string
}

// JobsRepository maintains list of job source code as blobs
type JobsRepository interface {
	Find(bprel.Job) (JobRecord, bool, error)
	Save(bprel.Job, JobRecord) error
}

type TemplateToJobRepository interface {
	FindByTemplate(bpdep.Template) (bprel.Job, bool, error)
	SaveForJob(bprel.Release, bprel.Job) error
}

type RuntimePackagesRepository interface {
	FindByReleaseJob(bprel.Job) ([]bprel.Package, bool, error)
	SaveForReleaseJob(bprel.Job, []bprel.Package) error

	// Keeps association between all possible packages for a job
	FindAllByReleaseJob(bprel.Job) ([]bprel.Package, bool, error)
	SaveAllForReleaseJob(bprel.Job, []bprel.Package) error
}
