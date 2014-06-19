package jobsrepo

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpindex "boshprovisioner/index"
	bprel "boshprovisioner/release"
)

type CRPRepository struct {
	index  bpindex.Index
	logger boshlog.Logger
}

func NewConcreteRuntimePackagesRepository(
	index bpindex.Index,
	logger boshlog.Logger,
) CRPRepository {
	return CRPRepository{
		index:  index,
		logger: logger,
	}
}

type jobToPkgsKey struct {
	// Mostly for ease of debugging
	JobName    string
	JobVersion string

	// Fingerprint of a job captures its dependenices
	JobFingerprint string

	// Indicates if contains all packages in a job's release
	ContainsAllPackages bool
}

func (r CRPRepository) FindByReleaseJob(job bprel.Job) ([]bprel.Package, bool, error) {
	var pkgs []bprel.Package

	err := r.index.Find(r.jobSpecificKey(job), &pkgs)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return pkgs, false, nil
		}

		return pkgs, false, bosherr.WrapError(err, "Finding rel-job -> specific rel-pkgs record")
	}

	return pkgs, true, nil
}

func (r CRPRepository) SaveForReleaseJob(job bprel.Job, pkgs []bprel.Package) error {
	err := r.index.Save(r.jobSpecificKey(job), pkgs)
	if err != nil {
		return bosherr.WrapError(err, "Saving rel-job -> specific rel-pkgs record")
	}

	return nil
}

// FindAllByReleaseJob keeps association between all possible packages for a job
func (r CRPRepository) FindAllByReleaseJob(job bprel.Job) ([]bprel.Package, bool, error) {
	var pkgs []bprel.Package

	err := r.index.Find(r.jobAllKey(job), &pkgs)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return pkgs, false, nil
		}

		return pkgs, false, bosherr.WrapError(err, "Finding rel-job -> all rel-pkgs record")
	}

	return pkgs, true, nil
}

func (r CRPRepository) SaveAllForReleaseJob(job bprel.Job, pkgs []bprel.Package) error {
	err := r.index.Save(r.jobAllKey(job), pkgs)
	if err != nil {
		return bosherr.WrapError(err, "Saving rel-job -> all rel-pkgs record")
	}

	return nil
}

func (r CRPRepository) jobSpecificKey(job bprel.Job) jobToPkgsKey {
	return jobToPkgsKey{
		JobName:        job.Name,
		JobVersion:     job.Version,
		JobFingerprint: job.Fingerprint,

		ContainsAllPackages: false,
	}
}

func (r CRPRepository) jobAllKey(job bprel.Job) jobToPkgsKey {
	return jobToPkgsKey{
		JobName:        job.Name,
		JobVersion:     job.Version,
		JobFingerprint: job.Fingerprint,

		ContainsAllPackages: true,
	}
}
