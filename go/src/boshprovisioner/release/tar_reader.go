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
	path       string
	downloader bpdload.Downloader
	extractor  bptar.Extractor
	fs         boshsys.FileSystem
	logger     boshlog.Logger

	// location to clean if successfully downloaded/extracted
	downloadPath string
	extractPath  string
}

func NewTarReader(
	path string,
	downloader bpdload.Downloader,
	extractor bptar.Extractor,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) *TarReader {
	return &TarReader{
		path:       path,
		downloader: downloader,
		extractor:  extractor,
		fs:         fs,
		logger:     logger,
	}
}

func (tr *TarReader) Read() (Release, error) {
	var release Release

	downloadPath, err := tr.downloader.Download(tr.path)
	if err != nil {
		return release, bosherr.WrapError(err, "Downloading release")
	}

	tr.downloadPath = downloadPath

	extractPath, err := tr.extractor.Extract(tr.downloadPath)
	if err != nil {
		cleanUpErr := tr.downloader.CleanUp(tr.downloadPath)
		if cleanUpErr != nil {
			tr.logger.Debug(tarReaderLogTag,
				"Failed to clean up downloaded release %v", cleanUpErr)
		}

		return release, bosherr.WrapError(err, "Extracting release")
	}

	tr.extractPath = extractPath

	manifestPath := filepath.Join(tr.extractPath, "release.MF")

	manifest, err := bprelman.NewManifestFromPath(manifestPath, tr.fs)
	if err != nil {
		closeErr := tr.Close()
		if closeErr != nil {
			tr.logger.Debug(tarReaderLogTag,
				"Failed to close release %v", closeErr)
		}

		return release, bosherr.WrapError(err, "Building manifest")
	}

	tr.logger.Debug(tarReaderLogTag, "Done building manifest %#v", manifest)

	release.populateFromManifest(manifest)

	tr.populateReleaseTarPaths(&release)

	return release, nil
}

func (tr TarReader) Close() error {
	dlErr := tr.downloader.CleanUp(tr.downloadPath)
	if dlErr != nil {
		tr.logger.Debug(tarReaderLogTag,
			"Failed to clean up downloaded release %v", dlErr)
	}

	exErr := tr.extractor.CleanUp(tr.extractPath)
	if exErr != nil {
		tr.logger.Debug(tarReaderLogTag,
			"Failed to clean up extracted release %v", exErr)
	}

	if dlErr != nil {
		return dlErr
	}

	return exErr
}

// populateReleaseTarPaths sets TarPath for each job/package in the release.
func (tr TarReader) populateReleaseTarPaths(release *Release) {
	for i, job := range release.Jobs {
		fileName := job.Name + ".tgz"
		release.Jobs[i].TarPath = filepath.Join(tr.extractPath, "jobs", fileName)
	}

	for _, pkg := range release.Packages {
		fileName := pkg.Name + ".tgz"
		pkg.TarPath = filepath.Join(tr.extractPath, "packages", fileName)
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
