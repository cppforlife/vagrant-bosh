package jobsrepo

import (
	"fmt"

	bpdep "boshprovisioner/deployment"
	bprel "boshprovisioner/release"
)

type JobRecord struct {
	BlobID string
	SHA1   string
}

type ReleaseJobRecord struct {
	ReleaseName    string
	ReleaseVersion string

	JobName        string
	JobVersion     string
	JobFingerprint string
}

func (r ReleaseJobRecord) String() string {
	return fmt.Sprintf("job %s in release %s/%s", r.JobName, r.ReleaseName, r.ReleaseVersion)
}

// JobsRepository maintains list of job source code as blobs
type JobsRepository interface {
	Find(bprel.Job) (JobRecord, bool, error)
	Save(bprel.Job, JobRecord) error

	FindByReleaseJob(ReleaseJobRecord) (JobRecord, bool, error)
}

type TemplateToJobRepository interface {
	FindByTemplate(bpdep.Template) (ReleaseJobRecord, bool, error)
	SaveForJob(bprel.Release, bprel.Job) (ReleaseJobRecord, error)
}

// RuntimePackagesRepository maintains list of releases' packages
type RuntimePackagesRepository interface {
	Find(ReleaseJobRecord) ([]bprel.Package, bool, error)
	Save(ReleaseJobRecord, []bprel.Package) error

	// Keeps association between all possible packages for a job
	FindAll(ReleaseJobRecord) ([]bprel.Package, bool, error)
	SaveAll(ReleaseJobRecord, []bprel.Package) error
}
