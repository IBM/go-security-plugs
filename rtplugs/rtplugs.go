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
//		rt := rtplugs.New(log)
//		if rt != nil {
//			defer rt.Close()
//			reverseproxy.Transport = rt.Transport(reverseproxy.Transport)
//		}
//
// While `log` is an optional logger
type RoundTrip struct {
	next           http.RoundTripper  // the next roundtripper
	roundTripPlugs []pi.RoundTripPlug // list of activated plugs
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

// New(pi.Logger) will attempt to strat a list of plugs
//
// env RTPLUGS defines a comma seperated list of plug names
// A typical RTPLUGS value would be "rtplug,wsplug"
// The plugs may be added statically (using imports) or dynmaicaly (.so files)
//
// For dynamically loaded plugs:
// The path of dynamicly included plugs should also be defined in RTPLUGS_SO
// env RTPLUGS_SO defines a comma seperated list of .so plug files
// relative/full path may be used
// A typical RTPLUGS_SO value would be "../../plugs/rtplug,../../plugs/wsplug"
// It is recommended to place the dynamic plugs in a plugs dir of the module.
// this helps ensure that plugs are built with the same package dependencies.
// Only plugs using the exact same package dependencies will be loaded.
func New(l pi.Logger) (rt *RoundTrip) {
	// Immidiatly return nil if RTPLUGS is not set
	plugsStr := os.Getenv("RTPLUGS")
	if plugsStr == "" {
		return
	}

	// Set logger for the entire RTPLUGS mechanism
	if l != nil {
		pi.Log = l
	}

	// Never panic the caller app from here
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during rtplugs.New()! One or more plugs may be skipped. Recover: %v", r)
		}
		if (rt != nil) && len(rt.roundTripPlugs) == 0 {
			rt = nil
		}
	}()

	// load any dynamic plugs
	load()

	plugs := strings.Split(plugsStr, ",")
	pi.Log.Infof("Trying to activate these %d plugs %v", len(plugs), plugs)

	for _, plugName := range plugs {
		for _, p := range pi.RoundTripPlugs {
			if p.PlugName() == plugName {
				// found a loaded plug, lets activate it
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
// existing RoundTripper will be screened using the security plugs
func (rt *RoundTrip) Transport(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = http.DefaultTransport
	}
	rt.next = t
	return rt
}

// Close() gracefully shuts down all plugs
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
	rt.roundTripPlugs = []pi.RoundTripPlug{}
}
