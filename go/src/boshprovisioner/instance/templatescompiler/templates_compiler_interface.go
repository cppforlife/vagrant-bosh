package templatescompiler

import (
	bpdep "boshprovisioner/deployment"
	bprel "boshprovisioner/release"
)

type RenderedArchiveRecord struct {
	SHA1   string
	BlobID string
}

type TemplatesCompiler interface {
	Precompile(bprel.Release) error
	Compile(bpdep.Job, bpdep.Instance) error
	FindRenderedArchive(bpdep.Job, bpdep.Instance) (RenderedArchiveRecord, error)

	// todo does it belong here?
	FindPackages(template bpdep.Template) ([]bprel.Package, error)
}
