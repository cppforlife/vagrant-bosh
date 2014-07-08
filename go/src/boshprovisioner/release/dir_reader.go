package release

import (
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bprelman "boshprovisioner/release/manifest"
)

const dirReaderLogTag = "DirReader"

type DirReader struct {
	releaseName    string // e.g. room101
	releaseVersion string // e.g. 0+dev.16
	dir            string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewDirReader(
	releaseName string,
	releaseVersion string,
	dir string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) DirReader {
	return DirReader{
		releaseName:    releaseName,
		releaseVersion: releaseVersion,
		dir:            dir,

		fs:     fs,
		logger: logger,
	}
}

func (r DirReader) Read() (Release, error) {
	var release Release

	// e.g. room101-0+dev.16.yml
	manifestFileName := r.releaseName + "-" + r.releaseVersion + ".yml"

	manifestPath := filepath.Join(r.dir, "dev_releases", manifestFileName)

	manifest, err := bprelman.NewManifestFromPath(manifestPath, r.fs)
	if err != nil {
		closeErr := r.Close()
		if closeErr != nil {
			r.logger.Debug(dirReaderLogTag,
				"Failed to close release %v", closeErr)
		}

		return release, bosherr.WrapError(err, "Building manifest")
	}

	r.logger.Debug(dirReaderLogTag, "Done building manifest %#v", manifest)

	release.populateFromManifest(manifest)

	r.populateReleaseTarPaths(&release)

	return release, nil
}

func (r DirReader) Close() error {
	// Caller owns release directory; hence, nothing to clean up
	return nil
}

// populateReleaseTarPaths sets TarPath for each job/package in the release.
func (r DirReader) populateReleaseTarPaths(release *Release) {
	devBuilds := filepath.Join(r.dir, ".dev_builds")
	finalBuilds := filepath.Join(r.dir, ".final_builds")

	for i, job := range release.Jobs {
		fileName := job.Fingerprint + ".tgz"

		devPath := filepath.Join(devBuilds, "jobs", job.Name, fileName)
		finalPath := filepath.Join(finalBuilds, "jobs", job.Name, fileName)

		release.Jobs[i].TarPath = r.pathThatExistsOrEmpty(devPath, finalPath)
	}

	for _, pkg := range release.Packages {
		fileName := pkg.Fingerprint + ".tgz"

		devPath := filepath.Join(devBuilds, "packages", pkg.Name, fileName)
		finalPath := filepath.Join(finalBuilds, "packages", pkg.Name, fileName)

		pkg.TarPath = r.pathThatExistsOrEmpty(devPath, finalPath)
	}
}

// pathThatExists returns first path that exists on the file system.
func (r DirReader) pathThatExistsOrEmpty(firstPath, secondPath string) string {
	switch {
	case r.fs.FileExists(firstPath):
		return firstPath
	case r.fs.FileExists(secondPath):
		return secondPath
	default:
		return ""
	}
}

/*
Example of BOSH release director with created dev releases:

$ tree ~/Downloads/room101-release
~/Downloads/room101-release
├── .dev_builds/
│ 	├── jobs
│ 	│   ├── warden
│ 	│   │   ├── 2a2b0559a97f869274602ffed008827cd66d15c3.tgz
│ 	│   │   └── index.yml
│ 	│   └── winston
│ 	│       ├── 98facd269cd8096a9b0ad354cbb5f0fc4265006f.tgz
│ 	│       └── index.yml
│ 	└── packages
│ 	    ├── aufs
│ 	    │   ├── cc5b6bf395c60d2aba6e0bc1ceeb613e7aadb52b.tgz
│ 	    │   └── index.yml
│ 	    ├── golang_1.2
│ 	    │   ├── ac825cab297fba938bec25c83f4a5780f88cdc92.tgz
│ 	    │   └── index.yml
│ 	    ├── iptables
│ 	    │   ├── 7226d311e90f49b05287e79f339581a1de9ea82e.tgz
│ 	    │   └── index.yml
│ 	    ├── pid_utils
│ 	    │   ├── de523512921338bac845ea7230e30b4307f842e7.tgz
│ 	    │   └── index.yml
│ 	    ├── warden-linux
│ 	    │   ├── 3f90138dae6c92c3e4595742ab6a513560e32a4c.tgz
│ 	    │   └── index.yml
│ 	    └── winston
│ 	        ├── 69fcbe3ef485a7f9d7c8efa6f18a74a0cffcb213.tgz
│ 	        └── index.yml
└── dev_releases
    ├── index.yml
    ├── room101-0+dev.1.yml
    └── room101-0+dev.2.yml
*/
