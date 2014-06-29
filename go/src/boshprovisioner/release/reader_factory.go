package release

import (
	"strings"

	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpdload "boshprovisioner/downloader"
	bptar "boshprovisioner/tar"
)

const (
	readerFactoryDirPrefix = "dir://"
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

func (rf ReaderFactory) NewReader(name, version, url string) Reader {
	if strings.HasPrefix(url, readerFactoryDirPrefix) {
		path := url[len(readerFactoryDirPrefix):]
		return NewDirReader(name, version, path, rf.fs, rf.logger)
	}

	return NewTarReader(url, rf.downloader, rf.extractor, rf.fs, rf.logger)
}
