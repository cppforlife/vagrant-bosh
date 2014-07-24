package templatescompiler

import (
	"fmt"

	boshblob "bosh/blobstore"
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpjobsrepo "boshprovisioner/instance/templatescompiler/jobsrepo"
	bptplsrepo "boshprovisioner/instance/templatescompiler/templatesrepo"
	bprel "boshprovisioner/release"
	bpreljob "boshprovisioner/release/job"
)

type ConcreteTemplatesCompiler struct {
	renderedArchivesCompiler RenderedArchivesCompiler
	jobReaderFactory         bpreljob.ReaderFactory

	jobsRepo      bpjobsrepo.JobsRepository
	tplToJobRepo  bpjobsrepo.TemplateToJobRepository
	runPkgsRepo   bpjobsrepo.RuntimePackagesRepository
	templatesRepo bptplsrepo.TemplatesRepository

	blobstore boshblob.Blobstore
	logger    boshlog.Logger
}

func NewConcreteTemplatesCompiler(
	renderedArchivesCompiler RenderedArchivesCompiler,
	jobReaderFactory bpreljob.ReaderFactory,
	jobsRepo bpjobsrepo.JobsRepository,
	tplToJobRepo bpjobsrepo.TemplateToJobRepository,
	runPkgsRepo bpjobsrepo.RuntimePackagesRepository,
	templatesRepo bptplsrepo.TemplatesRepository,
	blobstore boshblob.Blobstore,
	logger boshlog.Logger,
) ConcreteTemplatesCompiler {
	return ConcreteTemplatesCompiler{
		renderedArchivesCompiler: renderedArchivesCompiler,
		jobReaderFactory:         jobReaderFactory,

		jobsRepo:      jobsRepo,
		tplToJobRepo:  tplToJobRepo,
		runPkgsRepo:   runPkgsRepo,
		templatesRepo: templatesRepo,

		blobstore: blobstore,
		logger:    logger,
	}
}

// Precompile prepares release jobs to be later combined with instance properties
func (tc ConcreteTemplatesCompiler) Precompile(release bprel.Release) error {
	var allPkgs []bprel.Package

	for _, pkg := range release.Packages {
		if pkg == nil {
			// todo panic or should not be here?
			return bosherr.New("Expected release to not have nil package")
		}

		allPkgs = append(allPkgs, *pkg)
	}

	for _, job := range release.Jobs {
		jobRec, found, err := tc.jobsRepo.Find(job)
		if err != nil {
			return bosherr.WrapError(err, "Finding job source blob %s", job.Name)
		}

		if !found {
			blobID, fingerprint, err := tc.blobstore.Create(job.TarPath)
			if err != nil {
				return bosherr.WrapError(err, "Creating job source blob %s", job.Name)
			}

			jobRec = bpjobsrepo.JobRecord{
				BlobID: blobID,
				SHA1:   fingerprint,
			}

			err = tc.jobsRepo.Save(job, jobRec)
			if err != nil {
				return bosherr.WrapError(err, "Saving job record %s", job.Name)
			}
		}

		releaseJobRec, err := tc.tplToJobRepo.SaveForJob(release, job)
		if err != nil {
			return bosherr.WrapError(err, "Saving release job %s", job.Name)
		}

		// todo associate to release instead
		err = tc.runPkgsRepo.SaveAll(releaseJobRec, allPkgs)
		if err != nil {
			return bosherr.WrapError(err, "Saving release job %s", job.Name)
		}
	}

	return nil
}

// Compile populates blobstore with rendered jobs for a given deployment instance.
func (tc ConcreteTemplatesCompiler) Compile(job bpdep.Job, instance bpdep.Instance) error {
	blobID, fingerprint, err := tc.compileJob(job, instance)
	if err != nil {
		return err
	}

	templateRec := bptplsrepo.TemplateRecord{
		BlobID: blobID,
		SHA1:   fingerprint,
	}

	err = tc.templatesRepo.Save(job, instance, templateRec)
	if err != nil {
		return bosherr.WrapError(err, "Saving compiled templates record %s", job.Name)
	}

	return nil
}

// FindPackages returns list of packages required to run job template.
// List of packages is usually specified in release job metadata.
func (tc ConcreteTemplatesCompiler) FindPackages(template bpdep.Template) ([]bprel.Package, error) {
	var pkgs []bprel.Package

	releaseJobRec, found, err := tc.tplToJobRepo.FindByTemplate(template)
	if err != nil {
		return pkgs, bosherr.WrapError(err, "Finding release job record by template %s", template.Name)
	} else if !found {
		return pkgs, bosherr.New("Expected to find release job record by template %s", template.Name)
	}

	pkgs, found, err = tc.runPkgsRepo.Find(releaseJobRec)
	if err != nil {
		return pkgs, bosherr.WrapError(err, "Finding packages by release job record %v", releaseJobRec)
	} else if !found {
		return pkgs, bosherr.New("Expected to find packages by release job record %v", releaseJobRec)
	}

	return pkgs, nil
}

// FindRenderedArchive returns previously compiled template for a given instance.
// If such compiled template is not found, error is returned.
func (tc ConcreteTemplatesCompiler) FindRenderedArchive(job bpdep.Job, instance bpdep.Instance) (RenderedArchiveRecord, error) {
	var renderedArchiveRec RenderedArchiveRecord

	rec, found, err := tc.templatesRepo.Find(job, instance)
	if err != nil {
		return renderedArchiveRec, bosherr.WrapError(err, "Finding compiled templates %s", job.Name)
	} else if !found {
		return renderedArchiveRec, bosherr.New("Expected to find compiled templates %s", job.Name)
	}

	renderedArchiveRec.SHA1 = rec.SHA1
	renderedArchiveRec.BlobID = rec.BlobID

	return renderedArchiveRec, nil
}

// compileJob produces and saves rendered templates archive to a blobstore.
func (tc ConcreteTemplatesCompiler) compileJob(job bpdep.Job, instance bpdep.Instance) (string, string, error) {
	jobReaders, err := tc.buildJobReaders(job)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Building job readers")
	}

	var relJobs []bpreljob.Job

	for _, jobReader := range jobReaders {
		relJob, err := jobReader.tarReader.Read()
		if err != nil {
			return "", "", bosherr.WrapError(err, "Reading job")
		}

		defer jobReader.tarReader.Close()

		err = tc.associatePackages(jobReader.rec, relJob)
		if err != nil {
			return "", "", bosherr.WrapError(err, "Preparing runtime dep packages")
		}

		relJobs = append(relJobs, relJob)
	}

	renderedArchivePath, err := tc.renderedArchivesCompiler.Compile(relJobs, instance)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Compiling templates")
	}

	defer tc.renderedArchivesCompiler.CleanUp(renderedArchivePath)

	blobID, fingerprint, err := tc.blobstore.Create(renderedArchivePath)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Creating compiled templates")
	}

	return blobID, fingerprint, nil
}

type jobReader struct {
	rec       bpjobsrepo.ReleaseJobRecord
	tarReader bpreljob.Reader
}

func (tc ConcreteTemplatesCompiler) buildJobReaders(job bpdep.Job) ([]jobReader, error) {
	var readers []jobReader

	for _, template := range job.Templates {
		rec, found, err := tc.tplToJobRepo.FindByTemplate(template)
		if err != nil {
			return readers, bosherr.WrapError(err, "Finding dep-template -> release-job record %s", template.Name)
		} else if !found {
			return readers, bosherr.New("Expected to find dep-template -> release-job record %s", template.Name)
		}

		jobRec, found, err := tc.jobsRepo.FindByReleaseJob(rec)
		if err != nil {
			return readers, bosherr.WrapError(err, "Finding job source blob %s", template.Name)
		} else if !found {
			return readers, bosherr.New("Expected to find job source blob %s -- %s", template.Name, rec)
		}

		jobURL := fmt.Sprintf("blobstore:///%s?fingerprint=%s", jobRec.BlobID, jobRec.SHA1)

		reader := jobReader{
			rec:       rec,
			tarReader: tc.jobReaderFactory.NewReader(jobURL),
		}

		readers = append(readers, reader)
	}

	return readers, nil
}

func (tc ConcreteTemplatesCompiler) associatePackages(rec bpjobsrepo.ReleaseJobRecord, relJob bpreljob.Job) error {
	_, found, err := tc.runPkgsRepo.Find(rec)
	if err != nil {
		return bosherr.WrapError(err, "Finding runtime deps for %s", rec)
	}

	if found {
		return nil
	}

	// Find all packages in the same release,
	// regardless if job previously was associated with packages
	allPkgs, found, err := tc.runPkgsRepo.FindAll(rec)
	if err != nil {
		return bosherr.WrapError(err, "Finding rel-job -> rel-pkgs %s", rec)
	} else if !found {
		return bosherr.New("Expected to find rel-job -> rel-pkgs %s", rec)
	}

	var pkgs []bprel.Package

	// From all packages, select packages that are used by the job
	for _, pkg := range allPkgs {
		for _, p := range relJob.Packages {
			if pkg.Name == p.Name {
				pkgs = append(pkgs, pkg)
				break
			}
		}
	}

	// Return error if at least one depedency is missing
	if len(pkgs) != len(relJob.Packages) {
		return bosherr.New("Expected to find all release packages")
	}

	// Associate those packages with a job
	err = tc.runPkgsRepo.Save(rec, pkgs)
	if err != nil {
		return bosherr.WrapError(err, "Saving job packages %s", rec)
	}

	return nil
}
