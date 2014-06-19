package updater

import (
	"math"
	"time"

	boshaction "bosh/agent/action"
	boshas "bosh/agent/applier/applyspec"
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
)

const drainerLogTag = "Drainer"

type Drainer struct {
	agentClient bpagclient.Client
	logger      boshlog.Logger
}

func NewDrainer(
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) Drainer {
	return Drainer{
		agentClient: agentClient,
		logger:      logger,
	}
}

func (d Drainer) Drain() error {
	d.logger.Debug(drainerLogTag, "Draining instance")

	drainType := boshaction.DrainTypeUpdate
	spec := boshas.V1ApplySpec{}

	drainTime, err := d.agentClient.Drain(drainType, spec)
	if err != nil {
		return bosherr.WrapError(err, "Sending drain update")
	}

	if drainTime > 0 {
		d.logger.Debug(drainerLogTag, "Waiting for static drain to finish")
		time.Sleep(time.Duration(drainTime) * time.Second)
		return nil
	}

	d.logger.Debug(drainerLogTag, "Waiting for dynamic drain to finish")

	return d.waitForDynamicDrain(drainTime)
}

func (d Drainer) waitForDynamicDrain(drainTime int) error {
	var err error

	for {
		waitTime := int(math.Abs(float64(drainTime)))
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}

		if drainTime >= 0 {
			return nil
		}

		drainTime, err = d.agentClient.Drain(boshaction.DrainTypeStatus)
		if err != nil {
			return bosherr.WrapError(err, "Sending drain status")
		}
	}
}
