package release

import (
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bprelman "boshprovisioner/release/manifest"
)

const rawDirReaderLogTag = "DirReader"

type GitDirReader struct {
	dir    string
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewGitDirReader(
	dir string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) GitDirReader {
	return GitDirReader{
		dir:    dir,
		fs:     fs,
		logger: logger,
	}
}

func (r GitDirReader) Read() (Release, error) {
	var release Release

	var manifest bprelman.Manifest

	jobMatches, err := r.fs.Glob(filepath.Join(r.dir, "jobs/*"))
	if err != nil {
		return release, bosherr.WrapError(err, "Globbing jobs/ directory")
	}

	for _, jobMatch := range jobMatches {
		job := bprelman.Job{
			Name: filepath.Base(jobMatch),
		}
		manifest.Release.Jobs = append(manifest.Release.Jobs, job)
	}

	pkgMatches, err := r.fs.Glob(filepath.Join(r.dir, "packages/*"))
	if err != nil {
		return release, bosherr.WrapError(err, "Globbing packages/ directory")
	}

	for _, pkgMatch := range pkgMatches {
		pkg := bprelman.Package{
			Name: filepath.Base(pkgMatch),
		}
		manifest.Release.Packages = append(manifest.Release.Packages, pkg)
	}

	r.logger.Debug(dirReaderLogTag, "Done building manifest %s", manifest)

	release.populateFromManifest(manifest)

	r.populateReleaseDirPaths(&release)

	return release, nil
}

func (r GitDirReader) Close() error {
	// Caller owns release directory; hence, nothing to clean up
	return nil
}

// populateReleaseDirPaths sets Path for each job/package in the release.
func (r GitDirReader) populateReleaseDirPaths(release *Release) {
	for i, job := range release.Jobs {
		release.Jobs[i].TarPath = filepath.Join(r.dir, "jobs", job.Name)
	}

	for _, pkg := range release.Packages {
		pkg.TarPath = filepath.Join(r.dir, "packages", pkg.Name)
	}
}
