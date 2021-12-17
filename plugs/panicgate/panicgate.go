package main

import (
	"net/http"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

type plug struct {
	config map[string]interface{}
	log    pluginterfaces.Logger
}

var Plug plug = plug{}

func (plug) Initialize(l pluginterfaces.Logger, c map[string]interface{}) {
	Plug.log = l
	Plug.config = c
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in Initialize")
	}
	// dont panic so we get loaded
}

func (plug) Shutdown() {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in Shutdown")
	}
}

func (plug) PlugName() string {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in PlugName")
	}
	return "PanicGate"
}

func (plug) PlugVersion() string {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in PlugVersion")
	}
	return "my not so panicking version"
}

func (plug) PlugLogger() pluginterfaces.Logger {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in PlugLogger")
	}
	return Plug.log
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in ErrorHook")
	}

}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in ResponseHook")
	}
	if e := Plug.config["error"]; e != nil {
		return e.(error)
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	if Plug.config["panic"].(bool) {
		panic("it is fun to panic everywhere! also in RequestHook")
	}
	if e := Plug.config["error"]; e != nil {
		return e.(error)
	}
	return nil
}
