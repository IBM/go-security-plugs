package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"
const name string = "RoundTripGate"

type plug struct {
	name    string
	version string
	log     pluginterfaces.Logger
	// Add here any other state the extension needs
	config map[string]string
}

var Plug plug = plug{version: version, name: name}

func (plug) Initialize(l pluginterfaces.Logger) {
	Plug.log = l
	Plug.log.Infof("%s: Initializing - version %v\n", Plug.name, Plug.version)

	Plug.config = make(map[string]string)
	Plug.config["panicInitialize"] = os.Getenv("RT_GATE_PANIC_INIT")
	Plug.config["panicShutdown"] = os.Getenv("RT_GATE_PANIC_SHUTDOWN")
	Plug.config["panicReq"] = os.Getenv("RT_GATE_PANIC_REQ")
	Plug.config["panicResp"] = os.Getenv("RT_GATE_PANIC_RESP")
	Plug.config["errorReq"] = os.Getenv("RT_GATE_ERROR_REQ")
	Plug.config["errorResp"] = os.Getenv("RT_GATE_ERROR_RESP")

	if Plug.config["panicInitialize"] == "true" {
		panic("it is fun to panic everywhere! also in Initialize")
	}

}

func (plug) Shutdown() {
	Plug.log.Infof("%s: Shutdown", Plug.name)
	if Plug.config["panicShutdown"] == "true" {
		panic("it is fun to panic everywhere! also in Shutdown")
	}
}

func (plug) PlugName() string {
	return Plug.name
}

func (plug) PlugVersion() string {
	return Plug.version
}

func (plug) ApproveRequest(req *http.Request) (*http.Request, error) {
	Plug.log.Infof("%s: ApproveRequest started", Plug.name)
	if Plug.config["panicReq"] == "true" {
		panic("it is fun to panic everywhere! also in ApproveRequest")
	}

	if Plug.config["errorReq"] != "" {
		return nil, errors.New(Plug.config["errorReq"])
	}

	if req.Header.Get("X-Block-Req") != "" {
		Plug.log.Infof("%s ........... Blocked During Request! returning an error!", Plug.name)
		return nil, errors.New("request blocked")
	}

	for name, values := range req.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("%s Request Header: %s: %s", Plug.name, name, value)
		}
	}

	newCtx, cancelFunction := context.WithCancel(req.Context())
	req = req.WithContext(newCtx)

	timeoutStr := req.Header.Get("X-Block-Async")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeoutStr = "5s"
		timeout, _ = time.ParseDuration(timeoutStr)
	}
	Plug.log.Infof("%s ........... will asynchroniously block after %s", Plug.name, timeoutStr)

	go func(newCtx context.Context, cancelFunction context.CancelFunc, req *http.Request, timeout time.Duration) {
		select {
		case <-newCtx.Done():
			Plug.log.Infof("Done!")
		case <-time.After(timeout):
			Plug.log.Infof("Timeout!")
			cancelFunction()
		}
	}(newCtx, cancelFunction, req, timeout)

	return req, nil
}

func (plug) ApproveResponse(req *http.Request, resp *http.Response) error {
	Plug.log.Infof("%s: ApproveResponse started", Plug.name)
	if Plug.config["panicResp"] == "true" {
		panic("it is fun to panic everywhere! also in ApproveResponse")
	}

	if Plug.config["errorResp"] != "" {
		return errors.New(Plug.config["errorResp"])
	}

	if req.Header.Get("X-Block-Resp") != "" {
		Plug.log.Infof("%s ........... Blocked During Response! returning an error!", Plug.name)
		return errors.New("response blocked")
	}

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("%s Response Header: %s: %s", Plug.name, name, value)
		}
	}
	return nil
}
