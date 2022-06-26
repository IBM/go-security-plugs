package rtgate

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"
const name string = "rtgate"

type plug struct {
	name    string
	version string

	// Add here any other state the extension needs
	config map[string]string
}

func (p *plug) PlugName() string {
	return p.name
}

func (p *plug) PlugVersion() string {
	return p.version
}

func (p *plug) ApproveRequest(req *http.Request) (*http.Request, error) {
	pi.Log.Infof("%s: ApproveRequest started", p.name)
	pi.Log.Infof("Approve Request: panicReq %s", p.config["panicReq"])
	if p.config["panicReq"] == "true" {
		panic("it is fun to panic everywhere! also in ApproveRequest")
	}

	if p.config["errorReq"] != "" {
		return nil, errors.New(p.config["errorReq"])
	}

	if req.Header.Get("X-Block-Req") != "" {
		pi.Log.Infof("%s ........... Blocked During Request! returning an error!", p.name)
		return nil, errors.New("request blocked")
	}

	for name, values := range req.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Request Header: %s: %s", p.name, name, value)
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
	pi.Log.Infof("%s ........... will asynchroniously block after %s", p.name, timeoutStr)

	go func(newCtx context.Context, cancelFunction context.CancelFunc, req *http.Request, timeout time.Duration) {
		select {
		case <-newCtx.Done():
			pi.Log.Infof("Done!")
		case <-time.After(timeout):
			pi.Log.Infof("Timeout!")
			cancelFunction()
		}
	}(newCtx, cancelFunction, req, timeout)

	return req, nil
}

func (p *plug) ApproveResponse(req *http.Request, resp *http.Response) (*http.Response, error) {
	pi.Log.Infof("%s: ApproveResponse started", p.name)
	if p.config["panicResp"] == "true" {
		panic("it is fun to panic everywhere! also in ApproveResponse")
	}

	if p.config["errorResp"] != "" {
		return nil, errors.New(p.config["errorResp"])
	}

	if req.Header.Get("X-Block-Resp") != "" {
		pi.Log.Infof("%s ........... Blocked During Response! returning an error!", p.name)
		return nil, errors.New("response blocked")
	}

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Response Header: %s: %s", p.name, name, value)
		}
	}
	return resp, nil
}

func (p *plug) Shutdown() {
	pi.Log.Infof("%s: Shutdown", p.name)
	if p.config["panicShutdown"] == "true" {
		panic("it is fun to panic everywhere! also in Shutdown")
	}
}

func (p *plug) Init() {
	pi.Log.Infof("plug %s: Initializing - version %v", p.name, p.version)
	pi.Log.Infof("plug %s: Never use in production", p.name)
	p.config = make(map[string]string)
	p.config["panicInitialize"] = os.Getenv("RT_GATE_PANIC_INIT")
	p.config["panicShutdown"] = os.Getenv("RT_GATE_PANIC_SHUTDOWN")
	p.config["panicReq"] = os.Getenv("RT_GATE_PANIC_REQ")
	p.config["panicResp"] = os.Getenv("RT_GATE_PANIC_RESP")
	p.config["errorReq"] = os.Getenv("RT_GATE_ERROR_REQ")
	p.config["errorResp"] = os.Getenv("RT_GATE_ERROR_RESP")

	if p.config["panicInitialize"] == "true" {
		panic("it is fun to panic everywhere! also in Initialize")
	}
}

func init() {
	p := new(plug)
	p.version = version
	p.name = name
	pi.RegisterPlug(p)
}
