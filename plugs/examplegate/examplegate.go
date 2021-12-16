package main

import (
	"fmt"
	"net/http"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.7"

type plug struct {
	version string
	log     pluginterfaces.Logger
}

var Plug plug = plug{version: version}

func (plug) Initialize(l pluginterfaces.Logger) {
	fmt.Println("ExampleGate Initilizing... : ", Plug.log, l)
	Plug.log = l
	fmt.Println("ExampleGate Initilizing... : ", Plug.log, l)
	Plug.log.Infof("ExampleGate: Initializing - version %v\n", Plug.version)
}

func (plug) Shutdown() {
	Plug.log.Infof("ExampleGate: Shutdown")
}

func (plug) PlugName() string {
	return "ExampleGate"
}

func (plug) PlugVersion() string {
	return Plug.version
}

func (plug) PlugLogger() pluginterfaces.Logger {
	return Plug.log
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	Plug.log.Infof("ExampleGate: ErrorHook started")
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	Plug.log.Infof("ExampleGate: ResponseHook started")

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("ExampleGate Response Header: %s: %s", name, value)
		}
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	Plug.log.Infof("ExampleGate: RequestHook started")

	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			Plug.log.Infof("ExampleGate Request Header: %s: %s", name, value)
		}
	}
	return nil
}
