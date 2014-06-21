package provisioner

import (
	"fmt"

	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpeventlog "boshprovisioner/eventlog"
	bptplcomp "boshprovisioner/instance/templatescompiler"
	bppkgscomp "boshprovisioner/packagescompiler"
	bprel "boshprovisioner/release"
	bprelrepo "boshprovisioner/releasesrepo"
	bpvm "boshprovisioner/vm"
)

const releaseCompilerLogTag = "ReleaseCompiler"

type ReleaseCompiler struct {
	releasesRepo         bprelrepo.ReleasesRepository
	releaseReaderFactory bprel.ReaderFactory

	packagesCompilerFactory bppkgscomp.ConcretePackagesCompilerFactory
	templatesCompiler       bptplcomp.TemplatesCompiler

	vmProvisioner bpvm.VMProvisioner

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewReleaseCompiler(
	releasesRepo bprelrepo.ReleasesRepository,
	releaseReaderFactory bprel.ReaderFactory,
	packagesCompilerFactory bppkgscomp.ConcretePackagesCompilerFactory,
	templatesCompiler bptplcomp.TemplatesCompiler,
	vmProvisioner bpvm.VMProvisioner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) ReleaseCompiler {
	return ReleaseCompiler{
		releasesRepo:         releasesRepo,
		releaseReaderFactory: releaseReaderFactory,

		packagesCompilerFactory: packagesCompilerFactory,
		templatesCompiler:       templatesCompiler,

		vmProvisioner: vmProvisioner,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (p ReleaseCompiler) Compile(instance bpdep.Instance, releases []bpdep.Release) error {
	vm, err := p.vmProvisioner.Provision(instance)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning agent")
	}

	defer vm.Deprovision()

	err = p.uploadReleases(releases)
	if err != nil {
		return bosherr.WrapError(err, "Uploading releases")
	}

	pkgsCompiler := p.packagesCompilerFactory.NewCompiler(vm.AgentClient())

	err = p.compileReleases(pkgsCompiler, releases)
	if err != nil {
		return bosherr.WrapError(err, "Compiling releases")
	}

	return nil
}

func (p ReleaseCompiler) uploadReleases(releases []bpdep.Release) error {
	stage := p.eventLog.BeginStage("Uploading releases", len(releases)+1)

	for _, depRelease := range releases {
		releaseDesc := fmt.Sprintf("%s/%s", depRelease.Name, depRelease.Version)

		task := stage.BeginTask(fmt.Sprintf("Release %s", releaseDesc))

		err := task.End(p.releasesRepo.Pull(depRelease))
		if err != nil {
			return bosherr.WrapError(err, "Pulling release %s", depRelease.Name)
		}
	}

	task := stage.BeginTask("Deleting old releases")

	err := task.End(p.releasesRepo.KeepOnly(releases))
	if err != nil {
		return bosherr.WrapError(err, "Keeping only releases")
	}

	return nil
}

func (p ReleaseCompiler) compileReleases(pkgsCompiler bppkgscomp.PackagesCompiler, releases []bpdep.Release) error {
	for _, depRelease := range releases {
		err := p.compileRelease(pkgsCompiler, depRelease)
		if err != nil {
			return bosherr.WrapError(err, "Release %s", depRelease.Name)
		}
	}

	return nil
}

func (p ReleaseCompiler) compileRelease(pkgsCompiler bppkgscomp.PackagesCompiler, release bpdep.Release) error {
	reader := p.releaseReaderFactory.NewTarReader(release.URL)

	relRelease, err := reader.Read()
	if err != nil {
		return bosherr.WrapError(err, "Reading release")
	}

	defer reader.Close()

	err = pkgsCompiler.Compile(relRelease)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release packages")
	}

	err = p.templatesCompiler.Precompile(relRelease)
	if err != nil {
		return bosherr.WrapError(err, "Precompiling release job templates")
	}

	return nil
}
