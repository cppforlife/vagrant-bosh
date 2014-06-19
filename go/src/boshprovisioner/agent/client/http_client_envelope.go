package client

import (
	"encoding/json"

	boshaction "bosh/agent/action"
	bosherr "bosh/errors"
)

type requestEnvelope struct {
	Method    reqMethod `json:"method"`
	Arguments reqArgs   `json:"arguments"`
	ReplyTo   string    `json:"reply_to"`
}

type reqMethod string
type reqArgs []interface{}

type responseEnvelope struct {
	Value     json.RawMessage `json:"value"`
	Exception respException   `json:"exception"`
}

type respException struct {
	Message string `json:"message"`
}

type errandResultEnvelope struct {
	Result boshaction.ErrandResult `json:"result"`
}

type compiledPackageEnvelope struct {
	Result CompiledPackage `json:"result"`
}

func (re responseEnvelope) HasException() bool {
	return len(re.Exception.Message) > 0
}

func (re responseEnvelope) StringValue() (string, error) {
	value, err := re.interfaceValue()
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", bosherr.WrapError(err, "Converting value to string")
	}

	return str, nil
}

func (re responseEnvelope) MapValue() (map[string]interface{}, error) {
	var m map[string]interface{}

	value, err := re.interfaceValue()
	if err != nil {
		return m, err
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return m, bosherr.WrapError(err, "Converting value to map[string]interface{}")
	}

	return m, nil
}

func (re responseEnvelope) CustomValue(value interface{}) error {
	return json.Unmarshal(re.Value, &value)
}

func (re responseEnvelope) TaskID() (string, bool) {
	value, err := re.interfaceValue()
	if err != nil {
		return "", false
	}

	// Check if it looks like a task status
	result, ok := value.(map[string]interface{})
	if !ok {
		return "", false
	}

	taskID, found := result["agent_task_id"]
	if !found {
		return "", false
	}

	if result["state"] != "running" {
		return "", false
	}

	return taskID.(string), true
}

func (re responseEnvelope) interfaceValue() (interface{}, error) {
	var value interface{}

	err := json.Unmarshal(re.Value, &value)
	if err != nil {
		return value, err
	}

	return value, nil
}
