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
		rt.log.Debugf("Plug %s: ApproveRequest took %s", p.PlugName(), elapsed.String())
		if err != nil {
			rt.log.Infof("Plug %s: ApproveRequest returned an error %v", p.PlugName(), err)
			return
		}
	}
	return
}

func (rt *RoundTrip) approveResponse(req *http.Request, resp *http.Response) (err error) {
	for _, p := range rt.roudTripPlugs {
		start := time.Now()
		err = p.ApproveResponse(req, resp)
		elapsed := time.Since(start)
		rt.log.Debugf("Plug %s: ApproveResponse took %s", p.PlugName(), elapsed.String())
		if err != nil {
			rt.log.Infof("Plug %s: ApproveResponse returned an error %v", p.PlugName(), err)
			return
		}
	}
	return
}

func (rt *RoundTrip) nextRoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = rt.next.RoundTrip(req)
	elapsed := time.Since(start)
	rt.log.Debugf("Default DefaultTransport took %s\n", elapsed.String())
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

	req, err = rt.approveRequests(req)
	if err == nil {
		resp, err = rt.nextRoundTrip(req)
	}

	if err == nil {
		err = rt.approveResponse(req, resp)
	}
	if err != nil {
		rt.log.Infof("RoundTrip returned an error %v", err)
		resp = nil
	}
	return
}

func LoadPlugs(l pluginterfaces.Logger, plugins []string) (rt *RoundTrip) {
	rt = new(RoundTrip)

	if l == nil {
		l = dLog{}
	}

	rt.log = l
	rt.log.Infof("LoadPlugs started - trying these Plugins %v", plugins)

	defer func() {
		if r := recover(); r != nil {
			rt.log.Warnf("Recovered from panic during LoadPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		if (rt != nil) && len(rt.roudTripPlugs) == 0 {
			rt = nil
		}
	}()

	for _, plugPkgPath := range plugins {
		rt.log.Infof("Trying Plugin %s", plugPkgPath)

		plugPkg, err := plugin.Open(plugPkgPath)
		if err != nil {
			rt.log.Infof("Plugin %s skipped - Failed to open plugin. Err: %v", plugPkgPath, err)
			continue
		}

		if plugSymbol, err := plugPkg.Lookup("Plug"); err == nil {
			switch valType := plugSymbol.(type) {
			case pluginterfaces.RoundTripPlug:
				p := plugSymbol.(pluginterfaces.RoundTripPlug)
				p.Initialize(rt.log)
				rt.roudTripPlugs = append(rt.roudTripPlugs, p)
				rt.log.Infof("Plug %s (%s) was succesfully loaded", p.PlugName(), p.PlugVersion())

			default:
				rt.log.Infof("Plugin %s skipped - Plug symbol is of ilegal type %T,  %v", plugPkgPath, plugSymbol, valType)
			}

		} else {
			rt.log.Infof("Cant find Plug symbol in plugin: %s: %v", plugPkgPath, err)
			continue
		}
	}
	if len(rt.roudTripPlugs) == 0 {
		rt = nil
	}
	return
}

func Transport(rt *RoundTrip, t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = http.DefaultTransport
	}
	rt.next = t
	return rt
}

func UnloadPlugs(rt *RoundTrip) {
	if rt == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			rt.log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	for _, p := range rt.roudTripPlugs {
		p.Shutdown()
	}
}
