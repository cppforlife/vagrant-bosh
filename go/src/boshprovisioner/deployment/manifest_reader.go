package deployment

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpdepman "boshprovisioner/deployment/manifest"
)

type ManifestReader struct {
	path   string
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewManifestReader(
	path string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManifestReader {
	return ManifestReader{path: path, fs: fs, logger: logger}
}

func (r ManifestReader) Read() (Deployment, error) {
	var deployment Deployment

	manifest, err := bpdepman.NewManifestFromPath(r.path, r.fs)
	if err != nil {
		return deployment, bosherr.WrapError(err, "Reading manifest")
	}

	deployment.populateFromManifest(manifest)

	// todo pass by ref?
	err = NewSemanticValidator(deployment).Validate()
	if err != nil {
		return deployment, bosherr.WrapError(err, "Validating deployment semantically")
	}

	return deployment, nil
}

func (r ManifestReader) Close() error {
	return nil
}
