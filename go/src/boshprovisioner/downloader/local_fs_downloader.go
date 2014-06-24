package downloader

import (
	"strings"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

const localFSDownloaderLogTag = "LocalFSDownloader"

type LocalFSDownloader struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewLocalFSDownloader(
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) LocalFSDownloader {
	return LocalFSDownloader{fs: fs, logger: logger}
}

func (d LocalFSDownloader) Download(url string) (string, error) {
	file, err := d.fs.TempFile("downloader-LocalFSDownloader")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating download destination")
	}

	d.logger.Debug(localFSDownloaderLogTag, "Downloaded %s to %s", url, file.Name())

	err = file.Close()
	if err != nil {
		return "", bosherr.WrapError(err, "Closing download destination")
	}

	err = d.fs.CopyFile(strings.TrimPrefix(url, "file://"), file.Name())
	if err != nil {
		return "", bosherr.WrapError(err, "Copying to destination")
	}

	return file.Name(), nil
}

func (d LocalFSDownloader) CleanUp(path string) error {
	return d.fs.RemoveAll(path)
}
