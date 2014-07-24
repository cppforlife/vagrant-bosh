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

func (r *TarReader) Read() (Job, error) {
	var job Job

	downloadPath, err := r.downloader.Download(r.path)
	if err != nil {
		return job, bosherr.WrapError(err, "Downloading job")
	}

	r.downloadPath = downloadPath

	extractPath, err := r.extractor.Extract(r.downloadPath)
	if err != nil {
		cleanUpErr := r.downloader.CleanUp(r.downloadPath)
		if cleanUpErr != nil {
			r.logger.Debug(tarReaderLogTag,
				"Failed to clean up downloaded job %v", cleanUpErr)
		}

		return job, bosherr.WrapError(err, "Extracting job")
	}

	r.extractPath = extractPath

	manifestPath := filepath.Join(r.extractPath, "job.MF")

	manifest, err := bpreljobman.NewManifestFromPath(manifestPath, r.fs)
	if err != nil {
		closeErr := r.Close()
		if closeErr != nil {
			r.logger.Debug(tarReaderLogTag,
				"Failed to close job %v", closeErr)
		}

		return job, bosherr.WrapError(err, "Building manifest")
	}

	r.logger.Debug(tarReaderLogTag, "Done building manifest %#v", manifest)

	job.populateFromManifest(manifest)

	r.populateJobPaths(&job)

	return job, nil
}

func (r TarReader) Close() error {
	dlErr := r.downloader.CleanUp(r.downloadPath)
	if dlErr != nil {
		r.logger.Debug(tarReaderLogTag,
			"Failed to clean up downloaded job %v", dlErr)
	}

	exErr := r.extractor.CleanUp(r.extractPath)
	if exErr != nil {
		r.logger.Debug(tarReaderLogTag,
			"Failed to clean up extracted job %v", exErr)
	}

	if dlErr != nil {
		return dlErr
	}

	return exErr
}

// populateJobPaths sets Path for each template and monit in the job.
func (r TarReader) populateJobPaths(job *Job) {
	// monit file is outside of templates/ directory
	job.MonitTemplate.Path = filepath.Join(
		r.extractPath, job.MonitTemplate.SrcPathEnd)

	for i, template := range job.Templates {
		job.Templates[i].Path = filepath.Join(
			r.extractPath, "templates", template.SrcPathEnd)
	}
}

/*
Example layout of an unpackaged job tar:

$ tree ~/Downloads/dummy-job
~/Downloads/dummy-job
├── job.MF
├── monit
└── templates
    └── dummy_ctl.erb
*/
