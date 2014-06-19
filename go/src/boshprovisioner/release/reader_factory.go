package release

import (
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpdload "boshprovisioner/downloader"
	bptar "boshprovisioner/tar"
)

type ReaderFactory struct {
	downloader bpdload.Downloader
	extractor  bptar.Extractor
	fs         boshsys.FileSystem
	logger     boshlog.Logger
}

func NewReaderFactory(
	downloader bpdload.Downloader,
	extractor bptar.Extractor,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ReaderFactory {
	return ReaderFactory{
		downloader: downloader,
		extractor:  extractor,
		fs:         fs,
		logger:     logger,
	}
}

func (rf ReaderFactory) NewTarReader(path string) *TarReader {
	return NewTarReader(path, rf.downloader, rf.extractor, rf.fs, rf.logger)
}
