package packagescompiler

import (
	"fmt"

	boshcomp "bosh/agent/compiler"
	boshblob "bosh/blobstore"
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bragentclient "boshprovisioner/agent/client"
	bpeventlog "boshprovisioner/eventlog"
	bpcpkgsrepo "boshprovisioner/packagescompiler/compiledpackagesrepo"
	bppkgsrepo "boshprovisioner/packagescompiler/packagesrepo"
	bprel "boshprovisioner/release"
)

const concretePackagesCompilerLogTag = "ConcretePackagesCompiler"

type ConcretePackagesCompiler struct {
	agentClient          bragentclient.Client
	packagesRepo         bppkgsrepo.PackagesRepository
	compiledPackagesRepo bpcpkgsrepo.CompiledPackagesRepository
	blobstore            boshblob.Blobstore

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewConcretePackagesCompiler(
	agentClient bragentclient.Client,
	packagesRepo bppkgsrepo.PackagesRepository,
	compiledPackagesRepo bpcpkgsrepo.CompiledPackagesRepository,
	blobstore boshblob.Blobstore,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) ConcretePackagesCompiler {
	return ConcretePackagesCompiler{
		agentClient:          agentClient,
		packagesRepo:         packagesRepo,
		compiledPackagesRepo: compiledPackagesRepo,
		blobstore:            blobstore,

		eventLog: eventLog,
		logger:   logger,
	}
}

// Compile populates blobstore with compiled packages for a given release packages.
// All packages are compiled regardless if they will be later used or not.
// Currently Compile does not account for stemcell differences.
func (pc ConcretePackagesCompiler) Compile(release bprel.Release) error {
	packages := release.ResolvedPackageDependencies()

	releaseDesc := fmt.Sprintf("Compiling release %s/%s", release.Name, release.Version)

	stage := pc.eventLog.BeginStage(releaseDesc, len(packages))

	for _, pkg := range packages {
		pkgDesc := fmt.Sprintf("%s/%s", pkg.Name, pkg.Version)

		task := stage.BeginTask(fmt.Sprintf("Package %s", pkgDesc))

		_, found, err := pc.compiledPackagesRepo.Find(*pkg)
		if err != nil {
			return task.End(bosherr.WrapError(err, "Finding compiled package %s", pkg.Name))
		} else if found {
			task.End(nil)
			continue
		}

		err = task.End(pc.compilePkg(*pkg))
		if err != nil {
			return err
		}
	}

	return nil
}

// FindCompiledPackage returns previously compiled package for a given template.
// If such compiled package is not found, error is returned.
func (pc ConcretePackagesCompiler) FindCompiledPackage(pkg bprel.Package) (CompiledPackageRecord, error) {
	var compiledPkgRec CompiledPackageRecord

	rec, found, err := pc.compiledPackagesRepo.Find(pkg)
	if err != nil {
		return compiledPkgRec, bosherr.WrapError(err, "Finding compiled package %s", pkg.Name)
	} else if !found {
		return compiledPkgRec, bosherr.New("Expected to find compiled package %s", pkg.Name)
	}

	compiledPkgRec.SHA1 = rec.SHA1
	compiledPkgRec.BlobID = rec.BlobID

	return compiledPkgRec, nil
}

// compilePackage populates blobstore with a compiled package for a
// given package. Assumes that dependencies of given package have
// already been compiled and are in the blobstore.
func (pc ConcretePackagesCompiler) compilePkg(pkg bprel.Package) error {
	pc.logger.Debug(concretePackagesCompilerLogTag,
		"Preparing to compile package %v", pkg)

	pkgRec, found, err := pc.packagesRepo.Find(pkg)
	if err != nil {
		return bosherr.WrapError(err, "Finding package source blob %s", pkg.Name)
	}

	if !found {
		blobID, fingerprint, err := pc.blobstore.Create(pkg.TarPath)
		if err != nil {
			return bosherr.WrapError(err, "Creating package source blob %s", pkg.Name)
		}

		pkgRec = bppkgsrepo.PackageRecord{
			BlobID: blobID,
			SHA1:   fingerprint,
		}

		err = pc.packagesRepo.Save(pkg, pkgRec)
		if err != nil {
			return bosherr.WrapError(err, "Saving package record %s", pkg.Name)
		}
	}

	deps, err := pc.buildPkgDeps(pkg)
	if err != nil {
		return err
	}

	compiledPkgRes, err := pc.agentClient.CompilePackage(
		pkgRec.BlobID, // source tar
		pkgRec.SHA1,   // source tar
		pkg.Name,
		pkg.Version,
		deps,
	)
	if err != nil {
		return bosherr.WrapError(err, "Compiling package %s", pkg.Name)
	}

	compiledPkgRec := bpcpkgsrepo.CompiledPackageRecord{
		BlobID: compiledPkgRes.BlobID,
		SHA1:   compiledPkgRes.SHA1,
	}

	err = pc.compiledPackagesRepo.Save(pkg, compiledPkgRec)
	if err != nil {
		return bosherr.WrapError(err, "Saving compiled package %s", pkg.Name)
	}

	return nil
}

// buildPkgDeps prepares dependencies for agent's compile_package.
// Assumes that all package dependencies were already compiled.
func (pc ConcretePackagesCompiler) buildPkgDeps(pkg bprel.Package) (boshcomp.Dependencies, error) {
	deps := boshcomp.Dependencies{}

	for _, depPkg := range pkg.Dependencies {
		compiledPkgRec, found, err := pc.compiledPackagesRepo.Find(*depPkg)
		if err != nil {
			return deps, bosherr.WrapError(err, "Finding compiled package %s", depPkg.Name)
		} else if !found {
			return deps, bosherr.New("Expected to find compiled package %s", depPkg.Name)
		}

		deps[depPkg.Name] = boshcomp.Package{
			Name:        depPkg.Name,
			Version:     depPkg.Version,
			BlobstoreID: compiledPkgRec.BlobID, // compiled tar
			Sha1:        compiledPkgRec.SHA1,   // compiled tar
		}
	}

	return deps, nil
}
