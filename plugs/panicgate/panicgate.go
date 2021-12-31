package main

import (
	"errors"
	"net/http"
	"os"
	//pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

type plug struct {
	config map[string]string
	//log    pluginterfaces.Logger
}

var Plug plug = plug{}

func (plug) Initialize() {
	//Plug.log = l
	Plug.config = make(map[string]string)
	Plug.config["panicInitialize"] = os.Getenv("PANIC_GATE_PANIC_INIT")
	Plug.config["panicShutdown"] = os.Getenv("PANIC_GATE_PANIC_SHUTDOWN")
	Plug.config["panicReq"] = os.Getenv("PANIC_GATE_PANIC_REQ")
	Plug.config["panicResp"] = os.Getenv("PANIC_GATE_PANIC_RESP")
	Plug.config["panicErr"] = os.Getenv("PANIC_GATE_PANIC_ERR")
	Plug.config["error"] = os.Getenv("PANIC_GATE_ERROR")

	if Plug.config["panicInitialize"] == "true" {
		panic("it is fun to panic everywhere! also in Initialize")
	}
	// dont panic so we get loaded
}

func (plug) Shutdown() {
	if Plug.config["panicShutdown"] == "true" {
		panic("it is fun to panic everywhere! also in Shutdown")
	}
}

func (plug) PlugName() string {
	return "PanicGate"
}

func (plug) PlugVersion() string {
	return "my not so panicking version"
}

//ErrorHook(http.ResponseWriter, *http.Request, error)
func (plug) ErrorHook(w http.ResponseWriter, req *http.Request, e error) {
	if Plug.config["panicErr"] == "true" {
		panic("it is fun to panic everywhere! also in ErrorHook")
	}

}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	if Plug.config["panicResp"] == "true" {
		panic("it is fun to panic everywhere! also in ResponseHook")
	}
	if e := Plug.config["error"]; e != "" {
		return errors.New(e)
	}
	return nil
}

//RequestHook(http.ResponseWriter, *http.Request) error
func (plug) RequestHook(w http.ResponseWriter, r *http.Request) error {
	if Plug.config["panicRes"] == "true" {
		panic("it is fun to panic everywhere! also in RequestHook")
	}
	if e := Plug.config["error"]; e != "" {
		return errors.New(e)
	}
	return nil
}
