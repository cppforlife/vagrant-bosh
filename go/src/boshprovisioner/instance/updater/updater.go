package updater

import (
	"fmt"

	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpeventlog "boshprovisioner/eventlog"
	bpapplier "boshprovisioner/instance/updater/applier"
)

const updaterLogTag = "Updater"

type Updater struct {
	instanceDesc string

	preparer Preparer
	drainer  Drainer
	stopper  Stopper
	applier  bpapplier.Applier
	starter  Starter
	waiter   Waiter

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewUpdater(
	instanceDesc string,
	preparer Preparer,
	drainer Drainer,
	stopper Stopper,
	applier bpapplier.Applier,
	starter Starter,
	waiter Waiter,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) Updater {
	return Updater{
		instanceDesc: instanceDesc,

		preparer: preparer,
		drainer:  drainer,
		stopper:  stopper,
		applier:  applier,
		starter:  starter,
		waiter:   waiter,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (u Updater) Update() error {
	stage := u.eventLog.BeginStage(
		fmt.Sprintf("Updating instance %s", u.instanceDesc), 6)

	task := stage.BeginTask("Preparing")

	err := task.End(u.preparer.Prepare())
	if err != nil {
		return bosherr.WrapError(err, "Preparing")
	}

	task = stage.BeginTask("Draining")

	err = task.End(u.drainer.Drain())
	if err != nil {
		return bosherr.WrapError(err, "Draining")
	}

	task = stage.BeginTask("Stopping")

	err = task.End(u.stopper.Stop())
	if err != nil {
		return bosherr.WrapError(err, "Stopping")
	}

	task = stage.BeginTask("Applying")

	err = task.End(u.applier.Apply())
	if err != nil {
		return bosherr.WrapError(err, "Applying")
	}

	task = stage.BeginTask("Starting")

	err = task.End(u.starter.Start())
	if err != nil {
		return bosherr.WrapError(err, "Starting")
	}

	task = stage.BeginTask("Waiting")

	err = task.End(u.waiter.Wait())
	if err != nil {
		return bosherr.WrapError(err, "Waiting")
	}

	return nil
}
