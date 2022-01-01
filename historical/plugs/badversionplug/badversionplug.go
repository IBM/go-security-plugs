package main

import (
	"log"
	"net/http"
	"runtime"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"

type plug struct {
	version string
	config  map[string]interface{}
}

func init() {
	log.Println("This is BadVersionGate speaking: My go Version is ", runtime.Version())
}

var Plug plug = plug{version: version}

func (plug) Initialize(c map[string]interface{}) {
	Plug.config = c
	pi.Log.Infof("BadVersionGate: Initializing - version %v\n", Plug.version)
}

func (plug) Shutdown() {
	pi.Log.Infof("BadVersionGate: Shutdown")
}

func (plug) PlugName() string {
	return "BadVersionGate"
}

func (plug) PlugVersion() string {
	return Plug.version
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	pi.Log.Infof("BadVersionGate: ErrorHook started")
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	pi.Log.Infof("BadVersionGate: ResponseHook started")

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("BadVersionGate Response Header: %s: %s", name, value)
		}
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	pi.Log.Infof("BadVersionGate: RequestHook started")

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("BadVersionGate Request Header: %s: %s", name, value)
		}
	}
	return nil
}
