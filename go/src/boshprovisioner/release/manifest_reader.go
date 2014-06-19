package release

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
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

func (r ManifestReader) Read() (Release, error) {
	return Release{}, bosherr.New("Not implemented")
}

func (r ManifestReader) Close() error {
	return nil
}
