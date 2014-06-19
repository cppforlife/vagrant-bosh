package updater

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
)

const stopperLogTag = "Stopper"

type Stopper struct {
	agentClient bpagclient.Client
	logger      boshlog.Logger
}

func NewStopper(
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) Stopper {
	return Stopper{
		agentClient: agentClient,
		logger:      logger,
	}
}

func (s Stopper) Stop() error {
	s.logger.Debug(stopperLogTag, "Stopping instance")

	_, err := s.agentClient.Stop()
	if err != nil {
		return bosherr.WrapError(err, "Stopping")
	}

	return nil
}
