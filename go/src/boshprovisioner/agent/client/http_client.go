package client

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	boshaction "bosh/agent/action"
	boshas "bosh/agent/applier/applyspec"
	boshcomp "bosh/agent/compiler"
	bosherr "bosh/errors"
	boshlog "bosh/logger"
)

const httpClientLogTag = "HTTPClient"

type HTTPRequester interface {
	Do(*http.Request) (*http.Response, error)
}

type HTTPClient struct {
	url       *url.URL
	requester HTTPRequester
	logger    boshlog.Logger

	// Alias to method calls for convenience
	quickRequest requestFunc
	longRequest  requestFunc
}

type requestFunc func(reqMethod, reqArgs) (responseEnvelope, error)

func NewInsecureHTTPClientWithURI(uri string, logger boshlog.Logger) (HTTPClient, error) {
	mTLSConfig := &tls.Config{InsecureSkipVerify: true}
	transport := &http.Transport{TLSClientConfig: mTLSConfig}
	httpRequester := &http.Client{Transport: transport}
	return NewHTTPClientWithURI(uri, httpRequester, logger)
}

func NewHTTPClientWithURI(uri string, requester HTTPRequester, logger boshlog.Logger) (HTTPClient, error) {
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		return HTTPClient{}, bosherr.WrapError(err, "Parsing request uri")
	}

	return NewHTTPClient(url, requester, logger), nil
}

func NewHTTPClient(url *url.URL, requester HTTPRequester, logger boshlog.Logger) HTTPClient {
	client := HTTPClient{
		url:       url,
		requester: requester,
		logger:    logger,
	}

	client.quickRequest = client.makeQuickRequest
	client.longRequest = client.makeLongRequest

	return client
}

func (ac HTTPClient) Ping() (string, error) {
	return ac.makeStringRequest(ac.quickRequest, "ping", reqArgs{})
}

func (ac HTTPClient) GetTask(taskID string) (interface{}, error) {
	val, err := ac.makeQuickRequest("get_task", reqArgs{taskID})
	if err != nil {
		return nil, bosherr.WrapError(err, "makeRequest")
	}

	return val.MapValue()
}

func (ac HTTPClient) CancelTask(taskID string) (string, error) {
	return ac.makeStringRequest(ac.quickRequest, "cancel_task", reqArgs{taskID})
}

func (ac HTTPClient) SSH(cmd string, params boshaction.SshParams) (map[string]interface{}, error) {
	val, err := ac.makeQuickRequest("ssh", reqArgs{cmd, params})
	if err != nil {
		return nil, bosherr.WrapError(err, "makeRequest")
	}

	return val.MapValue()
}

func (ac HTTPClient) FetchLogs(logType string, filters []string) (map[string]interface{}, error) {
	val, err := ac.makeLongRequest("fetch_logs", reqArgs{logType, filters})
	if err != nil {
		return nil, bosherr.WrapError(err, "makeRequest")
	}

	return val.MapValue()
}

func (ac HTTPClient) Prepare(desiredSpec boshas.V1ApplySpec) (string, error) {
	return ac.makeStringRequest(ac.longRequest, "prepare", reqArgs{desiredSpec})
}

func (ac HTTPClient) Apply(desiredSpec boshas.V1ApplySpec) (string, error) {
	return ac.makeStringRequest(ac.longRequest, "apply", reqArgs{desiredSpec})
}

func (ac HTTPClient) Start() (string, error) {
	return ac.makeStringRequest(ac.quickRequest, "start", reqArgs{})
}

func (ac HTTPClient) Stop() (string, error) {
	return ac.makeStringRequest(ac.longRequest, "stop", reqArgs{})
}

func (ac HTTPClient) Drain(drainType boshaction.DrainType, newSpecs ...boshas.V1ApplySpec) (int, error) {
	var result int

	args := reqArgs{drainType}

	for _, newSpec := range newSpecs {
		args = append(args, newSpec)
	}

	val, err := ac.makeLongRequest("drain", args)
	if err != nil {
		return result, bosherr.WrapError(err, "makeRequest")
	}

	err = val.CustomValue(&result)
	if err != nil {
		return result, bosherr.WrapError(err, "Converting response to int")
	}

	return result, nil
}

func (ac HTTPClient) GetState(filters ...string) (boshaction.GetStateV1ApplySpec, error) {
	var result boshaction.GetStateV1ApplySpec

	args := reqArgs{}

	for _, filter := range filters {
		args = append(args, filter)
	}

	val, err := ac.makeQuickRequest("get_state", args)
	if err != nil {
		return result, bosherr.WrapError(err, "makeRequest")
	}

	err = val.CustomValue(&result)
	if err != nil {
		return result, bosherr.WrapError(err, "Converting response to apply spec")
	}

	return result, nil
}

func (ac HTTPClient) RunErrand() (boshaction.ErrandResult, error) {
	var result errandResultEnvelope

	val, err := ac.makeLongRequest("run_errand", reqArgs{})
	if err != nil {
		return result.Result, bosherr.WrapError(err, "makeRequest")
	}

	err = val.CustomValue(&result)
	if err != nil {
		return result.Result, bosherr.WrapError(
			err, "Converting response to errand result")
	}

	return result.Result, nil
}

func (ac HTTPClient) CompilePackage(blobID, sha1, name, version string, deps boshcomp.Dependencies) (CompiledPackage, error) {
	var result compiledPackageEnvelope

	val, err := ac.makeLongRequest(
		"compile_package",
		reqArgs{blobID, sha1, name, version, deps},
	)
	if err != nil {
		return result.Result, bosherr.WrapError(err, "makeRequest")
	}

	err = val.CustomValue(&result)
	if err != nil {
		return result.Result, bosherr.WrapError(
			err, "Converting response to errand result")
	}

	return result.Result, nil
}

func (ac HTTPClient) makeStringRequest(f requestFunc, method reqMethod, args reqArgs) (string, error) {
	val, err := f(method, args)
	if err != nil {
		return "", bosherr.WrapError(err, "makeXRequest")
	}

	return val.StringValue()
}

func (ac HTTPClient) makeLongRequest(method reqMethod, args reqArgs) (responseEnvelope, error) {
	val, err := ac.makeQuickRequest(method, args)
	if err != nil {
		return responseEnvelope{}, bosherr.WrapError(err, "makeRequest")
	}

	for {
		taskID, found := val.TaskID()
		if !found {
			return val, nil
		}

		time.Sleep(1 * time.Second)

		val, err = ac.makeQuickRequest("get_task", reqArgs{taskID})
		if err != nil {
			return responseEnvelope{}, bosherr.WrapError(err, "makeRequest")
		}
	}
}

func (ac HTTPClient) makeQuickRequest(method reqMethod, args reqArgs) (responseEnvelope, error) {
	var responseBody responseEnvelope

	requestBody := requestEnvelope{
		Method:    method,
		Arguments: args,
		ReplyTo:   "n-a",
	}

	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return responseBody, bosherr.WrapError(err, "Marshalling request body")
	}

	responseBytes, err := ac.makePlainRequest(string(requestBytes), "application/json")
	if err != nil {
		return responseBody, bosherr.WrapError(err, "Making plain request")
	}

	err = json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		return responseBody, bosherr.WrapError(err, "Unmarshalling response body")
	}

	if responseBody.HasException() {
		return responseBody, bosherr.New("Ended with exception %#v", responseBody.Exception)
	}

	return responseBody, nil
}

func (ac HTTPClient) makePlainRequest(requestBody, contentType string) ([]byte, error) {
	ac.logger.Debug(httpClientLogTag, "Making request url=%s", ac.url.String())

	ac.logger.DebugWithDetails(httpClientLogTag, "Request body", requestBody)

	request, err := http.NewRequest("POST", "", strings.NewReader(requestBody))
	if err != nil {
		return []byte{}, bosherr.WrapError(err, "Building request")
	}

	// Basic auth credentials are retrieved from the url
	request.URL = ac.url

	request.Header.Set("Content-Type", contentType)

	response, err := ac.requester.Do(request)
	if err != nil {
		if response != nil {
			ac.logger.Error(httpClientLogTag,
				"Received error=%v (response=%d)", err, response.StatusCode)
		} else {
			ac.logger.Error(httpClientLogTag,
				"Received error=%v (no response)", err)
		}
		return []byte{}, bosherr.WrapError(err, "Making request failed")
	}

	ac.logger.Debug(httpClientLogTag,
		"Received response status=%d", response.StatusCode)

	defer response.Body.Close()

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading response body")
	}

	ac.logger.DebugWithDetails(httpClientLogTag, "Response body", responseBytes)

	return responseBytes, nil
}
