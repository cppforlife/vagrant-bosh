package packagescompiler

import (
	bprel "boshprovisioner/release"
)

type CompiledPackageRecord struct {
	SHA1   string
	BlobID string
}

// PackagesCompiler takes each release package and compiles it.
// Compiled packages are used as:
//   (1) compile dependencies for other packages
//   (2) runtime dependencies for jobs
// todo account for stemcells
type PackagesCompiler interface {
	Compile(bprel.Release) error
	FindCompiledPackage(bprel.Package) (CompiledPackageRecord, error)
}
