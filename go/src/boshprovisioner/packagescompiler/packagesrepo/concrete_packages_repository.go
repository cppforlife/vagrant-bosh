package packagesrepo

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpindex "boshprovisioner/index"
	bprel "boshprovisioner/release"
)

type CPRepository struct {
	index  bpindex.Index
	logger boshlog.Logger
}

type pkgToPkgRecKey struct {
	// Mostly for ease of debugging
	PackageName    string
	PackageVersion string

	// Fingerprint of a package captures its dependenices
	PackageFingerprint string
}

func NewConcretePackagesRepository(
	index bpindex.Index,
	logger boshlog.Logger,
) CPRepository {
	return CPRepository{
		index:  index,
		logger: logger,
	}
}

func (r CPRepository) Find(pkg bprel.Package) (PackageRecord, bool, error) {
	var record PackageRecord

	err := r.index.Find(r.pkgKey(pkg), &record)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return record, false, nil
		}

		return record, false, bosherr.WrapError(err, "Finding package record")
	}

	return record, true, nil
}

func (r CPRepository) Save(pkg bprel.Package, record PackageRecord) error {
	err := r.index.Save(r.pkgKey(pkg), record)
	if err != nil {
		return bosherr.WrapError(err, "Saving package record")
	}

	return nil
}

func (r CPRepository) pkgKey(pkg bprel.Package) pkgToPkgRecKey {
	return pkgToPkgRecKey{
		PackageName:        pkg.Name,
		PackageVersion:     pkg.Version,
		PackageFingerprint: pkg.Fingerprint,
	}
}
