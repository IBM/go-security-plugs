package main

import (
	"errors"
	"net/http"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"
const name string = "ExampleGate"

type plug struct {
	name    string
	version string
	log     pluginterfaces.Logger
	config  map[string]interface{}
	// Add here any other state the extension needs
}

var Plug plug = plug{version: version, name: name}

func (plug) Initialize(l pluginterfaces.Logger, c map[string]interface{}) {
	Plug.log = l
	Plug.config = c
	Plug.log.Infof("%s: Initializing - version %v\n", Plug.name, Plug.version)
}

func (plug) Shutdown() {
	Plug.log.Infof("%s: Shutdown", Plug.name)
}

func (plug) PlugName() string {
	return Plug.name
}

func (plug) PlugVersion() string {
	return Plug.version
}

func (plug) PlugLogger() pluginterfaces.Logger {
	return Plug.log
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	Plug.log.Infof("%s: ErrorHook started", Plug.name)
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	Plug.log.Infof("%s: ResponseHook started", Plug.name)

	r := resp.Request
	if r.Header.Get("X-Block-Resp") != "" {
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

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	Plug.log.Infof("%s: RequestHook started", Plug.name)
	if r.Header.Get("X-Block-Req") != "" {
		Plug.log.Infof("%s ........... Blocked During Request! returning an error!", Plug.name)
		return errors.New("request blocked")
	}

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("%s Request Header: %s: %s", Plug.name, name, value)
		}
	}
	return nil
}
