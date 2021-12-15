package main

import (
	"net/http"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"

type plug struct {
	version string
	log     pluginterfaces.Logger
}

var Plug plug = plug{version: version}

func (p plug) Initialize(l pluginterfaces.Logger) {
	p.log = l
	p.log.Infof("ExampleGate: Initializing - version %v\n", p.version)
}

func (p plug) Shutdown() {
	p.log.Infof("ExampleGate: Shutdown")
}

func (p plug) PlugName() string {
	return "ExampleGate"
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (p plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	p.log.Infof("ExampleGate: ErrorHook started")
}

//ResponseHook(*http.Response) error
func (p plug) ResponseHook(resp *http.Response) error {
	p.log.Infof("ExampleGate: ResponseHook started")

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			p.log.Infof("ExampleGate Response Header: %s: %s", name, value)
		}
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (p plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	p.log.Infof("ExampleGate: RequestHook started")

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			p.log.Infof("ExampleGate Request Header: %s: %s", name, value)
		}
	}
	return nil
}
