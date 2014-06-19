package deployment

import (
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

type ReaderFactory struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewReaderFactory(
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ReaderFactory {
	return ReaderFactory{
		fs:     fs,
		logger: logger,
	}
}

func (rf ReaderFactory) NewManifestReader(path string) ManifestReader {
	return NewManifestReader(path, rf.fs, rf.logger)
}
