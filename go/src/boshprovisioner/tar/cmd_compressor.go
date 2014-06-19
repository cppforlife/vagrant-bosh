package tar

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

const cmdCompressorLogTag = "CmdCompressor"

type CmdCompressor struct {
	runner boshsys.CmdRunner
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewCmdCompressor(
	runner boshsys.CmdRunner,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) CmdCompressor {
	return CmdCompressor{
		runner: runner,
		fs:     fs,
		logger: logger,
	}
}

func (c CmdCompressor) Compress(path string) (string, error) {
	compressPath, err := c.tmpPath()
	if err != nil {
		return "", err
	}

	c.logger.Debug(cmdCompressorLogTag, "Compressing %s to %s", path, compressPath)

	_, _, _, err = c.runner.RunCommand("tar", "-C", path, "-czf", compressPath, ".")
	if err != nil {
		return "", bosherr.WrapError(err, "Running tar")
	}

	return compressPath, nil
}

func (c CmdCompressor) CleanUp(path string) error {
	return c.fs.RemoveAll(path)
}

func (c CmdCompressor) tmpPath() (string, error) {
	file, err := c.fs.TempFile("tar-CmdCompressor")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating extract destination")
	}

	compressPath := file.Name()

	err = file.Close()
	if err != nil {
		return "", bosherr.WrapError(err, "Closing temp file")
	}

	err = c.fs.RemoveAll(compressPath)
	if err != nil {
		return "", bosherr.WrapError(err, "Remove temp file")
	}

	return compressPath, nil
}
