package job

import (
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpdload "boshprovisioner/downloader"
	bpreljobman "boshprovisioner/release/job/manifest"
	bptar "boshprovisioner/tar"
)

const tarReaderLogTag = "TarReader"

// TarReader reads .tgz job file and returns a Job.
// See unpacked job directory layout at the end of the file.
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

func (tr *TarReader) Read() (Job, error) {
	var job Job

	downloadPath, err := tr.downloader.Download(tr.path)
	if err != nil {
		return job, bosherr.WrapError(err, "Downloading job")
	}

	tr.downloadPath = downloadPath

	extractPath, err := tr.extractor.Extract(tr.downloadPath)
	if err != nil {
		cleanUpErr := tr.downloader.CleanUp(tr.downloadPath)
		if cleanUpErr != nil {
			tr.logger.Debug(tarReaderLogTag,
				"Failed to clean up downloaded job %v", cleanUpErr)
		}

		return job, bosherr.WrapError(err, "Extracting job")
	}

	tr.extractPath = extractPath

	manifestPath := filepath.Join(tr.extractPath, "job.MF")

	manifest, err := bpreljobman.NewManifestFromPath(manifestPath, tr.fs)
	if err != nil {
		closeErr := tr.Close()
		if closeErr != nil {
			tr.logger.Debug(tarReaderLogTag,
				"Failed to close job %v", closeErr)
		}

		return job, bosherr.WrapError(err, "Building manifest")
	}

	tr.logger.Debug(tarReaderLogTag, "Done building manifest %#v", manifest)

	job.populateFromManifest(manifest)

	tr.populateJobPaths(&job)

	return job, nil
}

func (tr TarReader) Close() error {
	dlErr := tr.downloader.CleanUp(tr.downloadPath)
	if dlErr != nil {
		tr.logger.Debug(tarReaderLogTag,
			"Failed to clean up downloaded job %v", dlErr)
	}

	exErr := tr.extractor.CleanUp(tr.extractPath)
	if exErr != nil {
		tr.logger.Debug(tarReaderLogTag,
			"Failed to clean up extracted job %v", exErr)
	}

	if dlErr != nil {
		return dlErr
	}

	return exErr
}

// populateJobPaths sets Path for each template and monit in the job.
func (tr TarReader) populateJobPaths(job *Job) {
	// monit file is outside of templates/ directory
	job.MonitTemplate.Path = filepath.Join(
		tr.extractPath, job.MonitTemplate.SrcPathEnd)

	for i, template := range job.Templates {
		job.Templates[i].Path = filepath.Join(
			tr.extractPath, "templates", template.SrcPathEnd)
	}
}

/*
Example layout of an unpackaged job tar:

$ tree ~/Downloads/dummy-job
~/Downloads/dummy-job
├── job.MF
├── monit
└── templates
    └── dummy_ctl
*/
