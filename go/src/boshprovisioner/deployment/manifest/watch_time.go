package manifest

import (
	"strconv"
	"strings"

	bosherr "bosh/errors"
)

type WatchTime [2]int

func NewWatchTimeFromString(str string) (WatchTime, error) {
	var watchTime WatchTime

	parts := strings.Split(str, "-")
	if len(parts) != 2 {
		return watchTime, bosherr.New("Invalid watch time range %s", str)
	}

	min, err := strconv.Atoi(strings.Trim(parts[0], " "))
	if err != nil {
		return watchTime, bosherr.WrapError(
			err, "Non-positive number as watch time minimum %s", parts[0])
	}

	max, err := strconv.Atoi(strings.Trim(parts[1], " "))
	if err != nil {
		return watchTime, bosherr.WrapError(
			err, "Non-positive number as watch time maximum %s", parts[1])
	}

	if max < min {
		return watchTime, bosherr.New(
			"Watch time must have maximum greater than or equal minimum %s", str)
	}

	watchTime[0], watchTime[1] = min, max

	return watchTime, nil
}

func (wt WatchTime) Start() int { return wt[0] }
func (wt WatchTime) End() int   { return wt[1] }
