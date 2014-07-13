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
	ReleaseName    string
	ReleaseVersion string

	// Mostly for ease of debugging
	JobName    string
	JobVersion string

	// Fingerprint of a job captures its dependenices
	JobFingerprint string

	// Indicates if contains all packages in a job's release
	ContainsAllPackages bool
}

func (r CRPRepository) Find(rec ReleaseJobRecord) ([]bprel.Package, bool, error) {
	var pkgs []bprel.Package

	err := r.index.Find(r.jobSpecificKey(rec), &pkgs)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return pkgs, false, nil
		}

		return pkgs, false, bosherr.WrapError(err, "Finding rel-job record -> specific rel-pkgs record")
	}

	return pkgs, true, nil
}

func (r CRPRepository) Save(rec ReleaseJobRecord, pkgs []bprel.Package) error {
	err := r.index.Save(r.jobSpecificKey(rec), pkgs)
	if err != nil {
		return bosherr.WrapError(err, "Saving rel-job record -> specific rel-pkgs record")
	}

	return nil
}

func (r CRPRepository) FindAll(rec ReleaseJobRecord) ([]bprel.Package, bool, error) {
	var pkgs []bprel.Package

	err := r.index.Find(r.jobAllKey(rec), &pkgs)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return pkgs, false, nil
		}

		return pkgs, false, bosherr.WrapError(err, "Finding rel-job record -> all rel-pkgs record")
	}

	return pkgs, true, nil
}

func (r CRPRepository) SaveAll(rec ReleaseJobRecord, pkgs []bprel.Package) error {
	err := r.index.Save(r.jobAllKey(rec), pkgs)
	if err != nil {
		return bosherr.WrapError(err, "Saving rel-job record -> all rel-pkgs record")
	}

	return nil
}

func (r CRPRepository) jobSpecificKey(rec ReleaseJobRecord) jobToPkgsKey {
	return jobToPkgsKey{
		ReleaseName:    rec.ReleaseName,
		ReleaseVersion: rec.ReleaseVersion,

		JobName:        rec.JobName,
		JobVersion:     rec.JobVersion,
		JobFingerprint: rec.JobFingerprint,

		ContainsAllPackages: false,
	}
}

func (r CRPRepository) jobAllKey(rec ReleaseJobRecord) jobToPkgsKey {
	return jobToPkgsKey{
		ReleaseName:    rec.ReleaseName,
		ReleaseVersion: rec.ReleaseVersion,

		JobName:        rec.JobName,
		JobVersion:     rec.JobVersion,
		JobFingerprint: rec.JobFingerprint,

		ContainsAllPackages: true,
	}
}
