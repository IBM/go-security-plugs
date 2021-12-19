package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"

type plug struct {
	version string
	log     pluginterfaces.Logger
	config  map[string]interface{}
}

func init() {
	log.Println("This is BadVersionGate speaking: My go Version is ", runtime.Version())
}

var Plug plug = plug{version: version}

func (plug) Initialize(l pluginterfaces.Logger, c map[string]interface{}) {
	Plug.log = l
	Plug.config = c
	Plug.log.Infof("BadVersionGate: Initializing - version %v\n", Plug.version)
}

func (plug) Shutdown() {
	Plug.log.Infof("BadVersionGate: Shutdown")
}

func (plug) PlugName() string {
	return "BadVersionGate"
}

func (plug) PlugVersion() string {
	return Plug.version
}

func (plug) PlugLogger() pluginterfaces.Logger {
	return Plug.log
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	Plug.log.Infof("BadVersionGate: ErrorHook started")
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	Plug.log.Infof("BadVersionGate: ResponseHook started")

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("BadVersionGate Response Header: %s: %s", name, value)
		}
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	Plug.log.Infof("BadVersionGate: RequestHook started")

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("BadVersionGate Request Header: %s: %s", name, value)
		}
	}
	return nil
}
