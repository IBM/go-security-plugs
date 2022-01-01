package main

import (
	"errors"
	"net/http"
	"time"

	pi "github.com/IBM/go-security-plugs/historical/pluginterfaces"
)

const version string = "0.0.7"
const name string = "ProcessingGate"

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
	pi.Log.Infof("%s: ErrorHook", Plug.name)
}

//ResponseHook(*http.Response) error
func (plug) ResponseHook(resp *http.Response) error {
	pi.Log.Infof("%s: ResponseHook", Plug.name)

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
	pi.Log.Infof("%s: RequestHook", Plug.name)
	timeoutStr := r.Header.Get("X-Timeout")
	if timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			pi.Log.Infof("%s received request with errored X-Timeout header: %v", Plug.name, err)
			return nil
		}
		pi.Log.Infof("%s received request with X-Timout header: %s", Plug.name, timeoutStr)
		hj, ok := w.(http.Hijacker)
		if !ok {
			pi.Log.Errorf("%s doesn't support hijacking", Plug.name)
			return nil
		}

		go func(hj http.Hijacker, t time.Duration) {
			defer func() {
				if r := recover(); r != nil {
					pi.Log.Infof("%v", r)
				}

			}()
			time.Sleep(t)
			pi.Log.Infof("%s Hijacking due to %0.2f seconds Timeout!", Plug.name, float32(t)/1e9)
			conn, bufrw, err := hj.Hijack()
			if err != nil {
				pi.Log.Errorf("%s failed when trying to hijack: %v", Plug.name, err)
				return
			}
			// Don't forget to close the connection:
			bufrw.WriteString("Shalom, we dont want you here anymore!\n")
			bufrw.Flush()
			conn.Close()
			pi.Log.Infof("%s Closed Conn!!!!", Plug.name)
		}(hj, t)
		return nil
	}
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Request Header: %s: %s", Plug.name, name, value)
		}
	}
	return nil
}
