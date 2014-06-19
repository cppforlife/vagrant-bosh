package packagescompiler

import (
	boshblob "bosh/blobstore"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
	bpeventlog "boshprovisioner/eventlog"
	bpcpkgsrepo "boshprovisioner/packagescompiler/compiledpackagesrepo"
	bppkgsrepo "boshprovisioner/packagescompiler/packagesrepo"
)

type ConcretePackagesCompilerFactory struct {
	packagesRepo         bppkgsrepo.PackagesRepository
	compiledPackagesRepo bpcpkgsrepo.CompiledPackagesRepository
	blobstore            boshblob.Blobstore

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewConcretePackagesCompilerFactory(
	packagesRepo bppkgsrepo.PackagesRepository,
	compiledPackagesRepo bpcpkgsrepo.CompiledPackagesRepository,
	blobstore boshblob.Blobstore,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) ConcretePackagesCompilerFactory {
	return ConcretePackagesCompilerFactory{
		packagesRepo:         packagesRepo,
		compiledPackagesRepo: compiledPackagesRepo,
		blobstore:            blobstore,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (f ConcretePackagesCompilerFactory) NewCompiler(agentClient bpagclient.Client) PackagesCompiler {
	return NewConcretePackagesCompiler(
		agentClient,
		f.packagesRepo,
		f.compiledPackagesRepo,
		f.blobstore,
		f.eventLog,
		f.logger,
	)
}
