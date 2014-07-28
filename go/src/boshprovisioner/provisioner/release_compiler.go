package provisioner

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpeventlog "boshprovisioner/eventlog"
	bptplcomp "boshprovisioner/instance/templatescompiler"
	bppkgscomp "boshprovisioner/packagescompiler"
	bprel "boshprovisioner/release"
	bpvm "boshprovisioner/vm"
)

const releaseCompilerLogTag = "ReleaseCompiler"

type ReleaseCompiler struct {
	releaseReaderFactory bprel.ReaderFactory

	packagesCompilerFactory bppkgscomp.ConcretePackagesCompilerFactory
	templatesCompiler       bptplcomp.TemplatesCompiler

	vmProvisioner bpvm.Provisioner

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewReleaseCompiler(
	releaseReaderFactory bprel.ReaderFactory,
	packagesCompilerFactory bppkgscomp.ConcretePackagesCompilerFactory,
	templatesCompiler bptplcomp.TemplatesCompiler,
	vmProvisioner bpvm.Provisioner,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) ReleaseCompiler {
	return ReleaseCompiler{
		releaseReaderFactory: releaseReaderFactory,

		packagesCompilerFactory: packagesCompilerFactory,
		templatesCompiler:       templatesCompiler,

		vmProvisioner: vmProvisioner,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (p ReleaseCompiler) Compile(instance bpdep.Instance, depReleases []bpdep.Release) error {
	vm, err := p.vmProvisioner.Provision(instance)
	if err != nil {
		return bosherr.WrapError(err, "Provisioning VM")
	}

	defer vm.Deprovision()

	pkgsCompiler := p.packagesCompilerFactory.NewCompiler(vm.AgentClient())

	for _, depRelease := range depReleases {
		err := p.compileRelease(pkgsCompiler, depRelease)
		if err != nil {
			return bosherr.WrapError(err, "Release %s", depRelease.Name)
		}
	}

	return nil
}

func (p ReleaseCompiler) compileRelease(pkgsCompiler bppkgscomp.PackagesCompiler, depRelease bpdep.Release) error {
	relReader := p.releaseReaderFactory.NewReader(
		depRelease.Name,
		depRelease.Version,
		depRelease.URL,
	)

	relRelease, err := relReader.Read()
	if err != nil {
		return bosherr.WrapError(err, "Reading release")
	}

	defer relReader.Close()

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
