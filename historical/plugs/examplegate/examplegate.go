package main

import (
	"errors"
	"net/http"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"
const name string = "ExampleGate"

type plug struct {
	name    string
	version string
	// Add here any other state the extension needs
}

var Plug plug = plug{version: version, name: name}

func (plug) Initialize() {
	pi.Log.Infof("%s: Initializing - version %v\n", Plug.name, Plug.version)
}

func (plug) Shutdown() {
	pi.Log.Infof("%s: Shutdown", Plug.name)
}

func (plug) PlugName() string {
	return Plug.name
}

func (plug) PlugVersion() string {
	return Plug.version
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	pi.Log.Infof("%s: ErrorHook started", Plug.name)
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	pi.Log.Infof("%s: ResponseHook started", Plug.name)

	r := resp.Request
	if r.Header.Get("X-Block-Resp") != "" {
		pi.Log.Infof("%s ........... Blocked During Response! returning an error!", Plug.name)
		return errors.New("response blocked")
	}

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Response Header: %s: %s", Plug.name, name, value)
		}
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	pi.Log.Infof("%s: RequestHook started", Plug.name)
	if r.Header.Get("X-Block-Req") != "" {
		pi.Log.Infof("%s ........... Blocked During Request! returning an error!", Plug.name)
		return errors.New("request blocked")
	}

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Request Header: %s: %s", Plug.name, name, value)
		}
	}
	return nil
}
