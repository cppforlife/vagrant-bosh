package downloader

import (
	boshblob "bosh/blobstore"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

func NewDefaultMuxDownloader(
	blobstore boshblob.Blobstore,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) MuxDownloader {
	mux := map[string]Downloader{
		"http":      NewHTTPDownloader(fs, logger),
		"https":     NewHTTPDownloader(fs, logger),
		"file":      NewLocalFSDownloader(fs, logger),
		"blobstore": NewBlobstoreDownloader(blobstore, logger),
	}

	return NewMuxDownloader(mux, logger)
}
