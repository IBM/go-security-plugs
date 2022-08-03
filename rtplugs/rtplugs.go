// The rtplugs package instruments golang http clients that supports a RoundTripper interface.
// It was built and tested against https://pkg.go.dev/net/http/httputil#ReverseProxy
package rtplugs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"runtime/debug"
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
//
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
			pi.Log.Infof("rtplugs Plug %s: ApproveRequest returned an error %v", p.PlugName(), err)
			req = nil
			return
		}
		pi.Log.Debugf("rtplugs Plug %s: ApproveRequest took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) nextRoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = rt.next.RoundTrip(req)
	elapsed := time.Since(start)
	if err != nil {
		pi.Log.Infof("rtplugs nextRoundTrip (i.e. DefaultTransport) returned an error %v", err)
		resp = nil
		return
	}
	pi.Log.Debugf("rtplugs nextRoundTrip (i.e. DefaultTransport) took %s\n", elapsed.String())
	return
}

func (rt *RoundTrip) approveResponse(req *http.Request, respIn *http.Response) (resp *http.Response, err error) {
	resp = respIn
	for _, p := range rt.roundTripPlugs {
		start := time.Now()
		resp, err = p.ApproveResponse(req, resp)
		elapsed := time.Since(start)
		if err != nil {
			pi.Log.Infof("rtplugs Plug %s: ApproveResponse returned an error %v", p.PlugName(), err)
			resp = nil
			return
		}
		pi.Log.Debugf("rtplugs Plug %s: ApproveResponse took %s", p.PlugName(), elapsed.String())
	}
	return
}

func (rt *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			pi.Log.Warnf("rtplus Recovered from panic during RoundTrip! Recover: %v\n", recovered)
			pi.Log.Infof("rtplus stacktrace from panic: \n %s\n" + string(debug.Stack()))
			err = errors.New("paniced during RoundTrip")
			resp = nil
		}
	}()

	if req, err = rt.approveRequests(req); err == nil {
		if resp, err = rt.nextRoundTrip(req); err == nil {
			resp, err = rt.approveResponse(req, resp)
		}
	}
	return
}

// New(pi.Logger) will attempt to strat a list of plugs
//
// env RTPLUGS defines a comma seperated list of plug names
// A typical RTPLUGS value would be "rtplug,wsplug"
// The plugs may be added statically (using imports) or dynmaicaly (.so files)
func New(logger pi.Logger) (rt *RoundTrip) {
	pluglist := os.Getenv("RTPLUGS")
	if pluglist == "" {
		return
	}

	comma := func(c rune) bool {
		return c == ','
	}
	plugs := strings.FieldsFunc(pluglist, comma)

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		dat, err := os.ReadFile("/etc/podinfo/namespace")
		if err == nil {
			namespace = string(dat)
		}
	}
	svcname := os.Getenv("SERVICENAME")
	if svcname == "" {
		dat, err := os.ReadFile("/etc/podinfo/servicename")
		if err == nil {
			svcname = string(dat)
		}
	}

	if namespace == "" || svcname == "" {
		// mandatory
		panic("Can't find mandatory parameter namespace")
	}

	_, rt = NewConfigrablePlugs(context.Background(), logger, svcname, namespace, plugs, nil)
	return rt
}

// NewConfigrablePlugs(name, namespace, plugs, config, logger)
// Start `plugs` configured by `config`, to protect service `name` in `namespace`

// Context() wraps an existing RoundTripper
//
// Once the existing RoundTripper is wrapped, data flowing to and from the
// existing RoundTripper will be screened using the security plugs
func NewConfigrablePlugs(ctxin context.Context, logger pi.Logger, svcname string, namespace string, plugs []string, c map[string]map[string]string) (ctxout context.Context, rt *RoundTrip) {
	//skip for an empty pluglist
	if len(plugs) == 0 {
		return
	}

	// Set logger for the entire RTPLUGS mechanism
	if logger != nil {
		pi.Log = logger
	}

	// Never panic the caller app from here
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("rtplugs Recovered from panic during rtplugs.New()! One or more plugs may be skipped. Recover: %v", r)
		}
		if (rt != nil) && len(rt.roundTripPlugs) == 0 {
			rt = nil
		}
	}()

	ctxout = ctxin
	for _, plugName := range plugs {
		var foundPlug bool
		for _, p := range pi.RoundTripPlugs {
			if p.PlugName() == plugName {
				foundPlug = true
				var plugConfig map[string]string
				if c != nil {
					plugConfig = c[plugName]
				}
				// found a loaded plug, lets activate it
				pi.Log.Infof("Activating Plug %s with config %v", plugName, plugConfig)
				ctxout = p.Init(ctxout, plugConfig, svcname, namespace, logger)
				if rt == nil {
					rt = new(RoundTrip)
				}
				rt.roundTripPlugs = append(rt.roundTripPlugs, p)
				break
			}
		}
		if !foundPlug {
			pi.Log.Infof("Plug %s is not supported by this image. Consult your IT", plugName)
		}

	}
	for _, p := range rt.roundTripPlugs {
		pi.Log.Debugf("Plug %s version %s is active for service %s namespace %s", p.PlugName(), p.PlugVersion(), svcname, namespace)
	}
	return
}

// Transport() wraps an existing RoundTripper
//
// Once the existing RoundTripper is wrapped, data flowing to and from the
// existing RoundTripper will be screened using the security plugs
func (rt *RoundTrip) Transport(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		pi.Log.Debugf("Transport received a nil transport\n")
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
			pi.Log.Warnf("rtplugs Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		pi.Log.Sync()
	}()
	for _, p := range rt.roundTripPlugs {
		p.Shutdown()
	}
	rt.roundTripPlugs = []pi.RoundTripPlug{}
}
