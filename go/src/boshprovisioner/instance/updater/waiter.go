package updater

import (
	"time"

	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpagclient "boshprovisioner/agent/client"
)

const waiterLogTag = "Waiter"

var (
	ErrNotRunning = bosherr.New("Instance did not reach running state")
)

type Waiter struct {
	watchSchedule []time.Duration
	sleepFunc     SleepFunc
	agentClient   bpagclient.Client
	logger        boshlog.Logger
}

type SleepFunc func(time.Duration)

func NewWaiter(
	startWatchTime int,
	endWatchTime int,
	sleepFunc SleepFunc,
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) Waiter {
	return Waiter{
		watchSchedule: buildWatchSchedule(startWatchTime, endWatchTime),
		sleepFunc:     sleepFunc,
		agentClient:   agentClient,
		logger:        logger,
	}
}

// Wait waits for an instance to reach running state.
// It polls agent based on a watch schedule and inspects its job state.
func (w Waiter) Wait() error {
	w.logger.Debug(waiterLogTag, "Waiting for instance to reach running state")
	w.logger.Debug(waiterLogTag, "Using schedule %v", w.watchSchedule)

	for _, timeGap := range w.watchSchedule {
		w.logger.Debug(waiterLogTag, "Sleeping for %v", timeGap)
		w.sleepFunc(timeGap)

		state, err := w.agentClient.GetState()
		if err != nil {
			return bosherr.WrapError(err, "Sending get_state")
		}

		// todo stopped state
		if state.JobState == "running" {
			return nil
		}
	}

	return ErrNotRunning
}

// buildWatchSchedule returns list of millisecond interval
// at which instance should be checked for its running state.
// ([3000, 1000, 1000] =~ wait for 3000ms, then 1000ms, and 1000ms again)
func buildWatchSchedule(start, end int) []time.Duration {
	timeGaps := []time.Duration{time.Duration(start) * time.Millisecond}

	for total := start; total < end; total += 1000 {
		timeGaps = append(timeGaps, 1000*time.Millisecond)
	}

	return timeGaps
}
