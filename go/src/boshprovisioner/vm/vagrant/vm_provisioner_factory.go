package vagrant

import (
	boshlog "bosh/logger"
	boshsys "bosh/system"

	bpeventlog "boshprovisioner/eventlog"
	bpvm "boshprovisioner/vm"
)

type VMProvisionerFactory struct {
	fs     boshsys.FileSystem
	runner boshsys.CmdRunner

	assetsDir string
	mbus      string

	blobstoreConfig     map[string]interface{}
	vmProvisionerConfig bpvm.VMProvisionerConfig

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewVMProvisionerFactory(
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	assetsDir string,
	blobstoreConfig map[string]interface{},
	vmProvisionerConfig bpvm.VMProvisionerConfig,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) VMProvisionerFactory {
	return VMProvisionerFactory{
		fs:     fs,
		runner: runner,

		assetsDir:           assetsDir,
		blobstoreConfig:     blobstoreConfig,
		vmProvisionerConfig: vmProvisionerConfig,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (f VMProvisionerFactory) NewVMProvisioner() *VMProvisioner {
	cmds := NewSimpleCmds(f.runner, f.logger)

	vcapUserProvisioner := NewVCAPUserProvisioner(
		f.fs,
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
		f.vmProvisionerConfig.FullStemcellCompatibility,
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
		f.vmProvisionerConfig.AgentProvisioner,
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
