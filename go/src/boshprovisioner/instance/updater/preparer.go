package updater

import (
	boshas "bosh/agent/applier/applyspec"
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
)

const preparerLogTag = "Preparer"

type Preparer struct {
	agentClient bpagclient.Client
	logger      boshlog.Logger
}

func NewPreparer(
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) Preparer {
	return Preparer{
		agentClient: agentClient,
		logger:      logger,
	}
}

func (p Preparer) Prepare() error {
	p.logger.Debug(preparerLogTag, "Preparing instance")

	spec := boshas.V1ApplySpec{}

	_, err := p.agentClient.Prepare(spec)
	if err != nil {
		return bosherr.WrapError(err, "Sending prepare")
	}

	return nil
}
