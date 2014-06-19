package compiledpackagesrepo

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpindex "boshprovisioner/index"
	bprel "boshprovisioner/release"
)

type CCPRepository struct {
	index  bpindex.Index
	logger boshlog.Logger
}

type packageToCompiledPackageKey struct {
	PackageName    string
	PackageVersion string

	// Fingerprint of a package captures its dependenices
	PackageFingerprint string
}

func NewConcreteCompiledPackagesRepository(
	index bpindex.Index,
	logger boshlog.Logger,
) CCPRepository {
	return CCPRepository{index: index, logger: logger}
}

func (r CCPRepository) Find(pkg bprel.Package) (CompiledPackageRecord, bool, error) {
	var record CompiledPackageRecord

	err := r.index.Find(r.pkgKey(pkg), &record)
	if err != nil {
		if err == bpindex.ErrNotFound {
			return record, false, nil
		}

		return record, false, bosherr.WrapError(err, "Finding compiled package")
	}

	return record, true, nil
}

func (r CCPRepository) Save(pkg bprel.Package, record CompiledPackageRecord) error {
	err := r.index.Save(r.pkgKey(pkg), record)
	if err != nil {
		return bosherr.WrapError(err, "Saving compiled package")
	}

	return nil
}

func (r CCPRepository) pkgKey(pkg bprel.Package) packageToCompiledPackageKey {
	return packageToCompiledPackageKey{
		PackageName:        pkg.Name,
		PackageVersion:     pkg.Version,
		PackageFingerprint: pkg.Fingerprint,
	}
}
