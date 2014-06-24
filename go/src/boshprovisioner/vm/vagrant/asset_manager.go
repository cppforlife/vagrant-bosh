package vagrant

import (
	"fmt"
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

type AssetManager struct {
	rootDir string
	fs      boshsys.FileSystem
	runner  boshsys.CmdRunner
	logger  boshlog.Logger
}

func NewAssetManager(
	rootDir string,
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	logger boshlog.Logger,
) AssetManager {
	return AssetManager{
		rootDir: rootDir,
		fs:      fs,
		runner:  runner,
		logger:  logger,
	}
}

func (m AssetManager) Place(name, dstPath string) error {
	srcPath := filepath.Join(m.rootDir, name)

	if !m.fs.FileExists(srcPath) {
		return bosherr.New("Missing asset %s at %s", name, srcPath)
	}

	tempDir, err := m.fs.TempDir("vm-AssetManager")
	if err != nil {
		return bosherr.WrapError(err, "Creating temp dir")
	}

	defer m.fs.RemoveAll(tempDir)

	srcCopyPath := fmt.Sprintf("%s/copy", tempDir)

	// todo fs.CopyFile leaks fds
	_, _, _, err = m.runner.RunCommand("cp", srcPath, srcCopyPath)
	if err != nil {
		return bosherr.WrapError(err, "Copying asset")
	}

	err = m.fs.Rename(srcCopyPath, dstPath)
	if err != nil {
		return bosherr.WrapError(err, "Renaming asset %s to %s", name, dstPath)
	}

	return nil
}
