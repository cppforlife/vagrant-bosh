package erbrenderer

import (
	"encoding/json"
	"os"
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

const erbRendererLogTag = "ERBRenderer"

type ERBRenderer struct {
	fs      boshsys.FileSystem
	runner  boshsys.CmdRunner
	context TemplateEvaluationContext
	logger  boshlog.Logger

	rendererScript string
}

func NewERBRenderer(
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	context TemplateEvaluationContext,
	logger boshlog.Logger,
) ERBRenderer {
	return ERBRenderer{
		fs:      fs,
		runner:  runner,
		context: context,
		logger:  logger,

		rendererScript: templateEvaluationContextRb,
	}
}

func (r ERBRenderer) Render(srcPath, dstPath string) error {
	r.logger.Debug(erbRendererLogTag, "Rendering template %s", dstPath)

	dirPath := filepath.Dir(dstPath)

	err := r.fs.MkdirAll(dirPath, os.FileMode(0755))
	if err != nil {
		return bosherr.WrapError(err, "Creating directory %s", dirPath)
	}

	rendererScriptPath, err := r.writeRendererScript()
	if err != nil {
		return err
	}

	contextPath, err := r.writeContext()
	if err != nil {
		return err
	}

	// Use ruby to compile job templates
	command := boshsys.Command{
		Name: r.determineRubyExePath(),
		Args: []string{rendererScriptPath, contextPath, srcPath, dstPath},
	}

	_, _, _, err = r.runner.RunComplexCommand(command)
	if err != nil {
		return bosherr.WrapError(err, "Running ruby")
	}

	return nil
}

func (r ERBRenderer) writeRendererScript() (string, error) {
	// todo use temp path; it's same everytime?
	path := "/tmp/erb-render.rb"

	err := r.fs.WriteFileString(path, r.rendererScript)
	if err != nil {
		return "", bosherr.WrapError(err, "Writing renderer script")
	}

	return path, nil
}

func (r ERBRenderer) writeContext() (string, error) {
	contextBytes, err := json.Marshal(r.context)
	if err != nil {
		return "", bosherr.WrapError(err, "Marshalling context")
	}

	// todo use temp path?
	path := "/tmp/erb-context.json"

	err = r.fs.WriteFileString(path, string(contextBytes))
	if err != nil {
		return "", bosherr.WrapError(err, "Writing context")
	}

	return path, nil
}

func (r ERBRenderer) determineRubyExePath() string {
	// Prefer ruby executable on the PATH
	if r.runner.CommandExists("ruby") {
		return "ruby"
	}

	// Fallback to chef-solo ruby usually found in vagrant boxes
	vagrantRuby := "/opt/vagrant_ruby/bin/ruby"

	if r.fs.FileExists(vagrantRuby) {
		return vagrantRuby
	}

	return "ruby"
}
