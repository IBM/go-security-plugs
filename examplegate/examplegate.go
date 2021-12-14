package main

import (
	"net/http"

	"github.com/davidhadas/knativesecuritygate/reverseproxyplugs"
)

const version string = "0.0.6"

type plug struct {
	version string
}

var Plug plug
var log reverseproxyplugs.Logger

func init() {
	Plug.version = version
}

func (p plug) Initialize(l reverseproxyplugs.Logger) {
	log = l
	log.Infof("ExampleGate: Initializing - version %v\n", p.version)
}

func (plug) Shutdown() {
	log.Infof("ExampleGate: Shutdown")
}

func (plug) PlugName() string {
	return "ExampleGate"
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	log.Infof("ExampleGate: ErrorHook started")
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	log.Infof("ExampleGate: ResponseHook started")

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			log.Infof("ExampleGate Response Header: %s: %s", name, value)
		}
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	log.Infof("ExampleGate: RequestHook started")

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			log.Infof("ExampleGate Request Header: %s: %s", name, value)
		}
	}
	return nil
}
