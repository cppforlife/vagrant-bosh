package release

import (
	"strings"

	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpdload "boshprovisioner/downloader"
	bptar "boshprovisioner/tar"
)

const (
	readerFactoryDirPrefix    = "dir://"
	readerFactoryDirGitPrefix = "dir+git://"
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
	if strings.HasPrefix(url, readerFactoryDirGitPrefix) {
		dir := url[len(readerFactoryDirGitPrefix):]
		return NewGitDirReader(dir, rf.fs, rf.logger)
	}

	if strings.HasPrefix(url, readerFactoryDirPrefix) {
		dir := url[len(readerFactoryDirPrefix):]
		return NewDirReader(name, version, dir, rf.fs, rf.logger)
	}

	return NewTarReader(url, rf.downloader, rf.extractor, rf.fs, rf.logger)
}
