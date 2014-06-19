package packagesrepo

import (
	bprel "boshprovisioner/release"
)

type PackageRecord struct {
	BlobID string
	SHA1   string
}

// PackagesRepository maintains list of package source code as blobs.
type PackagesRepository interface {
	Find(bprel.Package) (PackageRecord, bool, error)
	Save(bprel.Package, PackageRecord) error
}
