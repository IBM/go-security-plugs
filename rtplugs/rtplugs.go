// The rtplugs package instruments golang http clients that supports a RoundTripper interface.
// It was built and tested against https://pkg.go.dev/net/http/httputil#ReverseProxy
package rtplugs

import (
	"errors"
	"net/http"
	"os"
	"strings"
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
	next           http.RoundTripper
	roundTripPlugs []pi.RoundTripPlug
}

func (rt *RoundTrip) approveRequests(reqin *http.Request) (req *http.Request, err error) {
	req = reqin
	for _, p := range rt.roundTripPlugs {
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
	for _, p := range rt.roundTripPlugs {
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

// New() will attempt to strat a list of plugins
//
// env RTPLUGS defines a comma seperated list of plugin names
// A typical RTPLUGS value would be "rtplug,wsplug"
// The plugins may be added statically (using imports) or dynmaicaly (.so files)
//
// For dynamic plugins:
// The path of dynamicly included plugins should also be defined in RTPLUGS_SO_PLUGINS
// env RTPLUGS_SO defines a comma seperated list of .so plugin files
// relative/full path may be used
// A typical RTPLUGS_SO value would be "../../plugs/rtplug,../../plugs/wsplug"
// It is recommended to place the dynamic plugins in a plugs dir of the module.
// this helps ensure that plugins are built with the same package dependencies.
// Only plugins using the exact same package dependencies will be loaded.
func New() (rt *RoundTrip) {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during rtplugs.New()! One or more plugs may be skipped. Recover: %v", r)
		}
		if (rt != nil) && len(rt.roundTripPlugs) == 0 {
			rt = nil
		}
	}()

	// load any dynamic plugins
	load()

	pluginsStr := os.Getenv("RTPLUGS")
	if pluginsStr == "" {
		return
	}

	plugins := strings.Split(pluginsStr, ",")
	pi.Log.Infof("Trying to activate these %d plugins %v", len(plugins), plugins)

	for _, plugName := range plugins {
		for _, p := range pi.RoundTripPlugs {
			if p.PlugName() == plugName {
				p.Init()
				if rt == nil {
					rt = new(RoundTrip)
				}
				rt.roundTripPlugs = append(rt.roundTripPlugs, p)
				pi.Log.Infof("Plugin %s is activated", plugName)
				break
			}
		}
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
	for _, p := range rt.roundTripPlugs {
		p.Shutdown()
	}
}
