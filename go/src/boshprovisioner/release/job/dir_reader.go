package job

import (
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpjobrelman "boshprovisioner/release/job/manifest"
)

const dirReaderLogTag = "DirReader"

type DirReader struct {
	dir string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewDirReader(
	dir string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) DirReader {
	return DirReader{
		dir: dir,

		fs:     fs,
		logger: logger,
	}
}

func (r DirReader) Read() (Job, error) {
	var job Job

	manifestPath := filepath.Join(r.dir, "spec")

	manifest, err := bpjobrelman.NewManifestFromPath(manifestPath, r.fs)
	if err != nil {
		closeErr := r.Close()
		if closeErr != nil {
			r.logger.Debug(dirReaderLogTag,
				"Failed to close job %v", closeErr)
		}

		return job, bosherr.WrapError(err, "Building manifest")
	}

	r.logger.Debug(dirReaderLogTag, "Done building manifest %#v", manifest)

	job.populateFromManifest(manifest)

	r.populateJobPaths(&job)

	return job, nil
}

func (r DirReader) Close() error {
	// Caller owns release directory; hence, nothing to clean up
	return nil
}

// populateJobPaths sets Path for each template and monit in the job.
func (r DirReader) populateJobPaths(job *Job) {
	// monit file is outside of templates/ directory
	job.MonitTemplate.Path = filepath.Join(r.dir, "monit")

	for i, template := range job.Templates {
		job.Templates[i].Path = filepath.Join(
			r.dir, "templates", template.SrcPathEnd)
	}
}

/*
Example layout of a job directory:

$ tree ~/Downloads/dummy-job
~/Downloads/dummy-job
├── spec
├── monit
└── templates
    └── dummy_ctl.erb
*/
