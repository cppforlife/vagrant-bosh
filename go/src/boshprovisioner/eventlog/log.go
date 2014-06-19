package eventlog

import (
	"encoding/json"
	"io"
	"os"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
)

const logLogTag = "Log"

type Log struct {
	writer io.Writer
	logger boshlog.Logger
}

type logEntry struct {
	Time int64 `json:"time"`

	Stage string   `json:"stage"`
	Task  string   `json:"task"`
	Tags  []string `json:"tags"`

	Total int `json:"total"`
	Index int `json:"index"`

	State    string `json:"state"`
	Progress int    `json:"progress"`

	// Might contain error key
	Data map[string]interface{} `json:"data,omitempty"`
}

func NewLog(logger boshlog.Logger) Log {
	return Log{
		writer: os.Stdout,
		logger: logger,
	}
}

func (l Log) BeginStage(name string, total int) *Stage {
	return &Stage{
		log:   l,
		name:  name,
		total: total,
	}
}

func (l Log) writeLogEntryNoErr(entry logEntry) {
	err := l.writeLogEntry(entry)
	if err != nil {
		l.logger.Error(logLogTag, "Failed writing log entry %s", err.Error())
	}
}

func (l Log) writeLogEntry(entry logEntry) error {
	bytes, err := json.Marshal(entry)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling log entry")
	}

	bytes = append(bytes, []byte("\n")...)

	_, err = l.writer.Write(bytes)
	if err != nil {
		return bosherr.WrapError(err, "Writing log entry")
	}

	return nil
}
