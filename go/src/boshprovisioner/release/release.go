package release

import (
	bprelman "boshprovisioner/release/manifest"
)

type Release struct {
	Manifest bprelman.Manifest

	Name    string
	Version string

	Jobs []Job

	Packages []*Package
}

type Job struct {
	Name    string
	Version string

	Fingerprint string
	SHA1        string

	TarPath string
}

type Package struct {
	Name    string
	Version string

	Fingerprint string
	SHA1        string

	TarPath string

	// Package dependencies used at compilation of this package
	Dependencies []*Package
}

// ResolvedPackageDependencies returns list of packages
// in order such that each package at a higher index
// only depends on packages at lower indecies.
func (r Release) ResolvedPackageDependencies() []*Package {
	var resolvedPkgs []*Package
	var pendingPkgs []*Package

	for _, pkg := range r.Packages {
		pkgRef := pkg
		pendingPkgs = append(pendingPkgs, pkgRef)
	}

	iterations := 0

	for len(pendingPkgs) > 0 {
		pkg := pendingPkgs[0]

		depsSatisfied := true

		for _, depPkg := range pkg.Dependencies {
			depSatisfied := false

			for _, resPkg := range resolvedPkgs {
				if depPkg.Name == resPkg.Name {
					depSatisfied = true
					break
				}
			}

			if !depSatisfied {
				depsSatisfied = false
				break
			}
		}

		pendingPkgs = pendingPkgs[1:]
		if depsSatisfied {
			resolvedPkgs = append(resolvedPkgs, pkg)
		} else {
			pendingPkgs = append(pendingPkgs, pkg)
		}

		if iterations > 100 {
			panic("Failed to resolve package dependenices witin 100 iterations")
		} else {
			iterations++
		}
	}

	return resolvedPkgs
}

// populateFromManifest populates release information
// interpreted from release manifest.
func (r *Release) populateFromManifest(manifest bprelman.Manifest) {
	r.populateRelease(manifest.Release)
	r.populatePackages(manifest.Release.Packages)
	r.populateJobs(manifest.Release.Jobs)
	r.Manifest = manifest
}

func (r *Release) populateRelease(manRelease bprelman.Release) {
	r.Name = manRelease.Name
	r.Version = manRelease.Version
}

func (r *Release) populatePackages(manPkgs []bprelman.Package) {
	nameToPkg := map[bprelman.DependencyName]*Package{}

	for _, manPkg := range manPkgs {
		pkg := Package{
			Name:    manPkg.Name,
			Version: manPkg.Version,

			Fingerprint: manPkg.Fingerprint,
			SHA1:        manPkg.SHA1,
		}

		r.Packages = append(r.Packages, &pkg)

		nameToPkg[bprelman.DependencyName(pkg.Name)] = &pkg
	}

	// Connect compile time dependencies for packages
	for i, manPkg := range manPkgs {
		for _, depName := range manPkg.DependencyNames {
			r.Packages[i].Dependencies = append(r.Packages[i].Dependencies, nameToPkg[depName])
		}
	}
}

func (r *Release) populateJobs(manJobs []bprelman.Job) {
	for _, manJob := range manJobs {
		job := Job{
			Name:    manJob.Name,
			Version: manJob.Version,

			Fingerprint: manJob.Fingerprint,
			SHA1:        manJob.SHA1,
		}

		r.Jobs = append(r.Jobs, job)
	}
}
