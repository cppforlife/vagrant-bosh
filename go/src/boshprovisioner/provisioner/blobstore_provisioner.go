package provisioner

import (
	"os"

	boshlog "bosh/logger"
	boshsys "bosh/system"
)

type BlobstoreProvisioner struct {
	fs              boshsys.FileSystem
	blobstoreConfig BlobstoreConfig
	logger          boshlog.Logger
}

func NewBlobstoreProvisioner(
	fs boshsys.FileSystem,
	blobstoreConfig BlobstoreConfig,
	logger boshlog.Logger,
) BlobstoreProvisioner {
	return BlobstoreProvisioner{
		fs:              fs,
		blobstoreConfig: blobstoreConfig,
		logger:          logger,
	}
}

func (p BlobstoreProvisioner) Provision() error {
	blobstorePath := p.blobstoreConfig.LocalPath()
	if blobstorePath != "" {
		err := p.fs.MkdirAll(blobstorePath, os.ModeDir)
		if err != nil {
			return err
		}
	}

	return nil
}
