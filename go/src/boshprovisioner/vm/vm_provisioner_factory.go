package vm

import (
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpeventlog "boshprovisioner/eventlog"
)

type VMProvisionerFactory struct {
	fs     boshsys.FileSystem
	runner boshsys.CmdRunner

	assetsDir       string
	mbus            string
	blobstoreConfig map[string]interface{}

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewVMProvisionerFactory(
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	assetsDir string,
	mbus string,
	blobstoreConfig map[string]interface{},
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) VMProvisionerFactory {
	return VMProvisionerFactory{
		fs:     fs,
		runner: runner,

		assetsDir:       assetsDir,
		mbus:            mbus,
		blobstoreConfig: blobstoreConfig,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (f VMProvisionerFactory) NewVMProvisioner() VMProvisioner {
	cmds := NewSimpleCmds(f.runner, f.logger)

	vcapUserProvisioner := NewVCAPUserProvisioner(
		cmds,
		f.runner,
		f.eventLog,
		f.logger,
	)

	assetManager := NewAssetManager(f.assetsDir, f.fs, f.runner, f.logger)

	runitProvisioner := NewRunitProvisioner(
		f.fs,
		cmds,
		f.runner,
		assetManager,
		f.logger,
	)

	monitProvisioner := NewMonitProvisioner(
		cmds,
		assetManager,
		runitProvisioner,
		f.logger,
	)

	depsProvisioner := NewDepsProvisioner(
		f.runner,
		f.eventLog,
		f.logger,
	)

	agentProvisioner := NewAgentProvisioner(
		f.fs,
		cmds,
		assetManager,
		runitProvisioner,
		monitProvisioner,
		f.blobstoreConfig,
		f.mbus,
		f.eventLog,
		f.logger,
	)

	vmProvisioner := NewVMProvisioner(
		vcapUserProvisioner,
		depsProvisioner,
		agentProvisioner,
		f.logger,
	)

	return vmProvisioner
}
