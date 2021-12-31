// The rtplugs package instruments golang http clients that supports a RoundTrip interface
// It was built and tested against the golang reverseproxy
//
// To extend reverseproxy use:
//		rt := rtplugs.New(logger, pluginList)
//		if roundTrip != nil {
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
	goLog "log"
	"net/http"
	"plugin"
	"time"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

type dLog struct{}

func (dLog) Debugf(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}
func (dLog) Infof(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}
func (dLog) Warnf(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}
func (dLog) Errorf(format string, args ...interface{}) {
	goLog.Printf(format, args...)
}

// An http.RoundTripper interface to be used as Transport for http clients
type RoundTrip struct {
	next          http.RoundTripper
	roudTripPlugs []pluginterfaces.RoundTripPlug
	log           pluginterfaces.Logger
}

func (rt *RoundTrip) approveRequests(reqin *http.Request) (req *http.Request, err error) {
	req = reqin
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		req, err = p.ApproveRequest(req)
		elapsed := time.Since(start)
		if err != nil {
			rt.log.Infof("Plug %s: ApproveRequest returned an error %v", p.PlugName(), err)
			req = nil
			return
		}
		rt.log.Debugf("Plug %s: ApproveRequest took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) nextRoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	rt.log.Debugf("nextRoundTrip rt.next.RoundTrip started\n")
	resp, err = rt.next.RoundTrip(req)
	rt.log.Debugf("nextRoundTrip rt.next.RoundTrip ended\n")
	elapsed := time.Since(start)
	if err != nil {
		rt.log.Infof("nextRoundTrip (i.e. DefaultTransport) returned an error %v", err)
		resp = nil
		return
	}
	rt.log.Debugf("nextRoundTrip (i.e. DefaultTransport) took %s\n", elapsed.String())
	return
}

func (rt *RoundTrip) approveResponse(req *http.Request, respIn *http.Response) (resp *http.Response, err error) {
	resp = respIn
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		resp, err = p.ApproveResponse(req, resp)
		elapsed := time.Since(start)
		if err != nil {
			rt.log.Infof("Plug %s: ApproveResponse returned an error %v", p.PlugName(), err)
			resp = nil
			return
		}
		rt.log.Debugf("Plug %s: ApproveResponse took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			rt.log.Warnf("Recovered from panic during RoundTrip! Recover: %v\n", recovered)
			err = errors.New("paniced during RoundTrip")
			resp = nil
		}
	}()

	if req, err = rt.approveRequests(req); err == nil {
		rt.log.Debugf("ApproveRequest ended")
		if resp, err = rt.nextRoundTrip(req); err == nil {
			rt.log.Debugf("nextRoundTrip ended")
			resp, err = rt.approveResponse(req, resp)
			rt.log.Debugf("approveResponse ended")
		}
	}
	return
}

// Use New() to load plugins while initializing or after calling Close()
// The plugins variable is a list of relative/full path to .so plugin files
// New() will attempt to load each of the plugins
// A good practice is to place the plugins in a plugs dir of the package,
// thereforea typical plugins value would be plugs = ["plugs/mygate/mygate.so"]
func New(l pluginterfaces.Logger, plugins []string) (rt *RoundTrip) {
	rt = new(RoundTrip)

	if l == nil {
		l = dLog{}
	}

	rt.log = l
	rt.log.Infof("LoadPlugs started - trying these Plugins %v", plugins)

	defer func() {
		if r := recover(); r != nil {
			rt.log.Warnf("Recovered from panic during New()!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		if (rt != nil) && len(rt.roudTripPlugs) == 0 {
			rt = nil
		}
	}()

	for _, plugPkgPath := range plugins {
		plugPkg, err := plugin.Open(plugPkgPath)
		if err != nil {
			rt.log.Warnf("Plugin %s skipped - failed to load so file. Err: %v", plugPkgPath, err)
			continue
		}

		newPlugSymbol, newPlugSymbolErr := plugPkg.Lookup("NewPlug")
		if newPlugSymbolErr != nil {
			rt.log.Warnf("Plugin %s skipped - missing 'NewPlug' symbol in plugin: %v", plugPkgPath, newPlugSymbolErr)
			continue
		}

		newPlug, newPlugTypeOk := newPlugSymbol.(func(pluginterfaces.Logger) pluginterfaces.RoundTripPlug)
		if !newPlugTypeOk {
			rt.log.Warnf("Plugin %s skipped - 'NewPlug' symbol is of ilegal type %T", plugPkgPath, newPlugSymbol)
			continue
		}
		// Okie Dokie - this plugin seems ok
		// Lets instantiate this new Plug
		p := newPlug(rt.log)

		rt.roudTripPlugs = append(rt.roudTripPlugs, p)

		rt.log.Infof("Plug %s (%s) was succesfully loaded", p.PlugName(), p.PlugVersion())
	}

	rt.log.Infof("Loaded plugs: %d - %v ", len(rt.roudTripPlugs), rt.roudTripPlugs)
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

// Use Close to gracefully shutdown plugs used
// Note that Close does not unload the .so files
// Instead, it informs all loaded plugs to gracefully shutdown and cleanup
func (rt *RoundTrip) Close() {
	defer func() {
		if r := recover(); r != nil {
			rt.log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	for _, p := range rt.roudTripPlugs {
		p.Shutdown()
	}
}
