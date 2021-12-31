// The rtplugs package instruments golang http clients that supports a RoundTrip interface
// It was built and tested against the golang reverseproxy
//
// To extend reverseproxy use:
//		rt := rtplugs.New(logger, pluginList)
//		if rt != nil {
//			defer rt.Close()
//			proxy.Transport = rt.Transport(proxy.Transport)
//		}
//
// While:
//    logger is the logger interface defined in package plugininterfaces
//    pluginList is a slice of strings for the path of plugins to load (.so files)
package rtplugs

import (
	"errors"
	"net/http"
	"plugin"
	"time"

	"github.com/IBM/go-security-plugs/pluginterfaces"
	"go.uber.org/zap"
)

// An http.RoundTripper interface to be used as Transport for http clients
type RoundTrip struct {
	next          http.RoundTripper
	roudTripPlugs []pluginterfaces.RoundTripPlug
	Log           pluginterfaces.Logger
}

func (rt *RoundTrip) approveRequests(reqin *http.Request) (req *http.Request, err error) {
	req = reqin
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		req, err = p.ApproveRequest(req)
		elapsed := time.Since(start)
		if err != nil {
			rt.Log.Infof("Plug %s: ApproveRequest returned an error %v", p.PlugName(), err)
			req = nil
			return
		}
		rt.Log.Debugf("Plug %s: ApproveRequest took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) nextRoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	rt.Log.Debugf("nextRoundTrip rt.next.RoundTrip started\n")
	resp, err = rt.next.RoundTrip(req)
	rt.Log.Debugf("nextRoundTrip rt.next.RoundTrip ended\n")
	elapsed := time.Since(start)
	if err != nil {
		rt.Log.Infof("nextRoundTrip (i.e. DefaultTransport) returned an error %v", err)
		resp = nil
		return
	}
	rt.Log.Debugf("nextRoundTrip (i.e. DefaultTransport) took %s\n", elapsed.String())
	return
}

func (rt *RoundTrip) approveResponse(req *http.Request, respIn *http.Response) (resp *http.Response, err error) {
	resp = respIn
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		resp, err = p.ApproveResponse(req, resp)
		elapsed := time.Since(start)
		if err != nil {
			rt.Log.Infof("Plug %s: ApproveResponse returned an error %v", p.PlugName(), err)
			resp = nil
			return
		}
		rt.Log.Debugf("Plug %s: ApproveResponse took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			rt.Log.Warnf("Recovered from panic during RoundTrip! Recover: %v\n", recovered)
			err = errors.New("paniced during RoundTrip")
			resp = nil
		}
	}()

	if req, err = rt.approveRequests(req); err == nil {
		rt.Log.Debugf("ApproveRequest ended")
		if resp, err = rt.nextRoundTrip(req); err == nil {
			rt.Log.Debugf("nextRoundTrip ended")
			resp, err = rt.approveResponse(req, resp)
			rt.Log.Debugf("approveResponse ended")
		}
	}
	return
}

// Use New() to load plugins while initializing or after calling Close()
// Providing the package with a logger allows
// The plugins variable is a list of relative/full path to .so plugin files
// New() will attempt to load each of the plugins
// A good practice is to place the plugins in a plugs dir of the package,
// thereforea typical plugins value would be plugs = ["plugs/mygate/mygate.so"]
func New(plugins []string) (rt *RoundTrip) {
	rt = new(RoundTrip)
	rt.SetLogger(nil)
	rt.Log.Infof("LoadPlugs started - trying these Plugins %v", plugins)

	defer func() {
		if r := recover(); r != nil {
			rt.Log.Warnf("Recovered from panic during New()!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		if (rt != nil) && len(rt.roudTripPlugs) == 0 {
			rt = nil
		}
	}()

	for _, plugPkgPath := range plugins {
		plugPkg, err := plugin.Open(plugPkgPath)
		if err != nil {
			rt.Log.Warnf("Plugin %s skipped - failed to load so file. Err: %v", plugPkgPath, err)
			continue
		}

		newPlugSymbol, newPlugSymbolErr := plugPkg.Lookup("NewPlug")
		if newPlugSymbolErr != nil {
			rt.Log.Warnf("Plugin %s skipped - missing 'NewPlug' symbol in plugin: %v", plugPkgPath, newPlugSymbolErr)
			continue
		}

		newPlug, newPlugTypeOk := newPlugSymbol.(func(pluginterfaces.Logger) pluginterfaces.RoundTripPlug)
		if !newPlugTypeOk {
			rt.Log.Warnf("Plugin %s skipped - 'NewPlug' symbol is of ilegal type %T", plugPkgPath, newPlugSymbol)
			continue
		}
		// Okie Dokie - this plugin seems ok
		// Lets instantiate this new Plug
		p := newPlug(rt.Log)

		rt.roudTripPlugs = append(rt.roudTripPlugs, p)

		rt.Log.Infof("Plug %s (%s) was succesfully loaded", p.PlugName(), p.PlugVersion())
	}

	rt.Log.Infof("Loaded plugs: %d - %v ", len(rt.roudTripPlugs), rt.roudTripPlugs)
	if len(rt.roudTripPlugs) == 0 {
		rt = nil
	}
	return
}

// Use Transport to add the loaded plugins to the chain of RoundTrippers used
func (rt *RoundTrip) Transport(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = http.DefaultTransport
	}
	rt.next = t
	return rt
}

// Set an alternative logger that meets the pluginterfaces.Logger interface
func (rt *RoundTrip) SetLogger(l pluginterfaces.Logger) {
	if l == nil {
		// set the default logger
		logger, _ := zap.NewProduction()
		rt.Log = logger.Sugar()
		return
	}
	rt.Log = l
}

// Use Close to gracefully shutdown plugs used
// Note that Close does not unload the .so files
// Instead, it informs all loaded plugs to gracefully shutdown and cleanup
func (rt *RoundTrip) Close() {
	defer func() {
		if r := recover(); r != nil {
			rt.Log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		rt.Log.Sync()
	}()
	for _, p := range rt.roudTripPlugs {
		p.Shutdown()
	}
}
