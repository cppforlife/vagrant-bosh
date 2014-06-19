package downloader

import (
	gourl "net/url"

	boshblob "bosh/blobstore"
	bosherr "bosh/errors"
	boshlog "bosh/logger"
)

const blobstoreDownloaderLogTag = "BlobstoreDownloader"

type BlobstoreDownloader struct {
	blobstore boshblob.Blobstore
	logger    boshlog.Logger
}

func NewBlobstoreDownloader(
	blobstore boshblob.Blobstore,
	logger boshlog.Logger,
) BlobstoreDownloader {
	return BlobstoreDownloader{
		blobstore: blobstore,
		logger:    logger,
	}
}

// Download takes URL of format blobstore:///blobId?fingerprint=sha1-value
func (d BlobstoreDownloader) Download(url string) (string, error) {
	parsedURL, err := gourl.Parse(url)
	if err != nil {
		return "", bosherr.WrapError(err, "Parsing url %s", url)
	}

	var fingerprint string

	if fingerprints, found := parsedURL.Query()["fingerprint"]; found {
		if len(fingerprints) > 0 {
			fingerprint = fingerprints[0]
		}
	}

	path, err := d.blobstore.Get(parsedURL.Path, fingerprint)
	if err != nil {
		return "", bosherr.WrapError(err, "Downloading blob")
	}

	d.logger.Debug(blobstoreDownloaderLogTag, "Downloaded %s to %s", url, path)

	return path, nil
}

func (d BlobstoreDownloader) CleanUp(path string) error {
	return d.blobstore.CleanUp(path)
}
