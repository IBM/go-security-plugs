// The rtplugs package instruments golang http clients that supports a RoundTripper interface.
// It was built and tested against https://pkg.go.dev/net/http/httputil#ReverseProxy
package rtplugs

import (
	"errors"
	"net/http"
	"plugin"
	"time"

	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

// An http.RoundTripper interface to be used as Transport for http clients
//
// To extend reverseproxy use:
//		rt := rtplugs.New(pluginList)
//		if rt != nil {
//			defer rt.Close()
//			reverseproxy.Transport = rt.Transport(reverseproxy.Transport)
//		}
//
// While `pluginList` is a slice of strings for the path of plugins (.so files) to load
type RoundTrip struct {
	next          http.RoundTripper
	roudTripPlugs []pi.RoundTripPlug
}

func (rt *RoundTrip) approveRequests(reqin *http.Request) (req *http.Request, err error) {
	req = reqin
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		req, err = p.ApproveRequest(req)
		elapsed := time.Since(start)
		if err != nil {
			pi.Log.Infof("Plug %s: ApproveRequest returned an error %v", p.PlugName(), err)
			req = nil
			return
		}
		pi.Log.Debugf("Plug %s: ApproveRequest took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) nextRoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	pi.Log.Debugf("nextRoundTrip rt.next.RoundTrip started\n")
	resp, err = rt.next.RoundTrip(req)
	pi.Log.Debugf("nextRoundTrip rt.next.RoundTrip ended\n")
	elapsed := time.Since(start)
	if err != nil {
		pi.Log.Infof("nextRoundTrip (i.e. DefaultTransport) returned an error %v", err)
		resp = nil
		return
	}
	pi.Log.Debugf("nextRoundTrip (i.e. DefaultTransport) took %s\n", elapsed.String())
	return
}

func (rt *RoundTrip) approveResponse(req *http.Request, respIn *http.Response) (resp *http.Response, err error) {
	resp = respIn
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		resp, err = p.ApproveResponse(req, resp)
		elapsed := time.Since(start)
		if err != nil {
			pi.Log.Infof("Plug %s: ApproveResponse returned an error %v", p.PlugName(), err)
			resp = nil
			return
		}
		pi.Log.Debugf("Plug %s: ApproveResponse took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			pi.Log.Warnf("Recovered from panic during RoundTrip! Recover: %v\n", recovered)
			err = errors.New("paniced during RoundTrip")
			resp = nil
		}
	}()

	if req, err = rt.approveRequests(req); err == nil {
		pi.Log.Debugf("ApproveRequest ended")
		if resp, err = rt.nextRoundTrip(req); err == nil {
			pi.Log.Debugf("nextRoundTrip ended")
			resp, err = rt.approveResponse(req, resp)
			pi.Log.Debugf("approveResponse ended")
		}
	}
	return
}

// New() will attempt to load each of the plugins
//
// `plugins` is a list of relative/full path to .so plugin files.
//
// Load plugins while initializing ( or after calling `Close()`)
//
// It is recommended to place the plugins in a plugs dir of the module.
// this help ensure that plugins are built with the same package dependencies.
// Only plugins the same package dependencies will be loaded.
//
// A typical plugins value would be plugs = ["plugs/mygate/mygate.so"]
func New(plugins []string) (rt *RoundTrip) {
	rt = new(RoundTrip)
	pi.Log.Infof("LoadPlugs started - trying these Plugins %v", plugins)

	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during New()!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		if (rt != nil) && len(rt.roudTripPlugs) == 0 {
			rt = nil
		}
	}()

	for _, plugPkgPath := range plugins {
		plugPkg, err := plugin.Open(plugPkgPath)
		if err != nil {
			pi.Log.Warnf("Plugin %s skipped - failed to load so file. Err: %v", plugPkgPath, err)
			continue
		}

		newPlugSymbol, newPlugSymbolErr := plugPkg.Lookup("NewPlug")
		if newPlugSymbolErr != nil {
			pi.Log.Warnf("Plugin %s skipped - missing 'NewPlug' symbol in plugin: %v", plugPkgPath, newPlugSymbolErr)
			continue
		}

		newPlug, newPlugTypeOk := newPlugSymbol.(func() pi.RoundTripPlug)
		if !newPlugTypeOk {
			pi.Log.Warnf("Plugin %s skipped - 'NewPlug' symbol is of ilegal type %T", plugPkgPath, newPlugSymbol)
			continue
		}
		// Okie Dokie - this plugin seems ok
		// Lets instantiate this new Plug
		p := newPlug()

		rt.roudTripPlugs = append(rt.roudTripPlugs, p)

		pi.Log.Infof("Plug %s (%s) was succesfully loaded", p.PlugName(), p.PlugVersion())
	}

	pi.Log.Infof("Loaded plugs: %d - %v ", len(rt.roudTripPlugs), rt.roudTripPlugs)
	if len(rt.roudTripPlugs) == 0 {
		rt = nil
	}
	return
}

// Transport() wraps an existing RoundTripper
//
// Once the existing RoundTripper is wrapped, data flowing to and from the
// existing RoundTripper will be screened using the security plugins
func (rt *RoundTrip) Transport(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = http.DefaultTransport
	}
	rt.next = t
	return rt
}

// Close() gracefully shuts down all plugins
//
// Note that Close does not unload the .so files,
// instead, it informs all loaded plugs to gracefully shutdown and cleanup
func (rt *RoundTrip) Close() {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		pi.Log.Sync()
	}()
	for _, p := range rt.roudTripPlugs {
		p.Shutdown()
	}
}
