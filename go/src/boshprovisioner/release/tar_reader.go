package release

import (
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpdload "boshprovisioner/downloader"
	bprelman "boshprovisioner/release/manifest"
	bptar "boshprovisioner/tar"
)

const tarReaderLogTag = "TarReader"

// TarReader reads .tgz release file and returns a Release.
// See unpacked release directory layout at the end of the file.
type TarReader struct {
	url string

	downloader bpdload.Downloader
	extractor  bptar.Extractor
	fs         boshsys.FileSystem
	logger     boshlog.Logger

	// location to clean if successfully downloaded/extracted
	downloadPath string
	extractPath  string
}

func NewTarReader(
	url string,
	downloader bpdload.Downloader,
	extractor bptar.Extractor,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) *TarReader {
	return &TarReader{
		url: url,

		downloader: downloader,
		extractor:  extractor,
		fs:         fs,
		logger:     logger,
	}
}

func (r *TarReader) Read() (Release, error) {
	var release Release

	downloadPath, err := r.downloader.Download(r.url)
	if err != nil {
		return release, bosherr.WrapError(err, "Downloading release")
	}

	r.downloadPath = downloadPath

	extractPath, err := r.extractor.Extract(r.downloadPath)
	if err != nil {
		cleanUpErr := r.downloader.CleanUp(r.downloadPath)
		if cleanUpErr != nil {
			r.logger.Debug(tarReaderLogTag,
				"Failed to clean up downloaded release %v", cleanUpErr)
		}

		return release, bosherr.WrapError(err, "Extracting release")
	}

	r.extractPath = extractPath

	manifestPath := filepath.Join(r.extractPath, "release.MF")

	manifest, err := bprelman.NewManifestFromPath(manifestPath, r.fs)
	if err != nil {
		closeErr := r.Close()
		if closeErr != nil {
			r.logger.Debug(tarReaderLogTag,
				"Failed to close release %v", closeErr)
		}

		return release, bosherr.WrapError(err, "Building manifest")
	}

	r.logger.Debug(tarReaderLogTag, "Done building manifest %#v", manifest)

	release.populateFromManifest(manifest)

	r.populateReleaseTarPaths(&release)

	return release, nil
}

func (r TarReader) Close() error {
	dlErr := r.downloader.CleanUp(r.downloadPath)
	if dlErr != nil {
		r.logger.Debug(tarReaderLogTag,
			"Failed to clean up downloaded release %v", dlErr)
	}

	exErr := r.extractor.CleanUp(r.extractPath)
	if exErr != nil {
		r.logger.Debug(tarReaderLogTag,
			"Failed to clean up extracted release %v", exErr)
	}

	if dlErr != nil {
		return dlErr
	}

	return exErr
}

// populateReleaseTarPaths sets TarPath for each job/package in the release.
func (r TarReader) populateReleaseTarPaths(release *Release) {
	for i, job := range release.Jobs {
		fileName := job.Name + ".tgz"
		release.Jobs[i].TarPath = filepath.Join(r.extractPath, "jobs", fileName)
	}

	for _, pkg := range release.Packages {
		fileName := pkg.Name + ".tgz"
		pkg.TarPath = filepath.Join(r.extractPath, "packages", fileName)
	}
}

/*
Example layout of an unpackaged release tar:

$ tree ~/Downloads/dummy-release
~/Downloads/dummy-release
├── jobs
│   ├── dummy.tgz
│   ├── dummy_with_bad_package.tgz
│   ├── dummy_with_package.tgz
│   └── dummy_with_properties.tgz
├── packages
│   ├── bad_package.tgz
│   └── dummy_package.tgz
└── release.MF
*/
