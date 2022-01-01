package reverseproxyplugs

import (
	"errors"
	"net/http"
	"plugin"
	"time"

	pi "github.com/IBM/go-security-plugs/historical/pluginterfaces"
)

var reverseProxyPlugs []pi.ReverseProxyPlug

func init() {
	initialize()
}

func initialize() {
	reverseProxyPlugs = []pi.ReverseProxyPlug{}
	//log = dLog{}
}

func LoadPlugs(plugins []string) (ret int) {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during LoadPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	ret = len(reverseProxyPlugs)

	//if l != nil {
	//	log = l
	//}
	for _, plugPkgPath := range plugins {
		plugPkg, err := plugin.Open(plugPkgPath)
		if err != nil {
			pi.Log.Infof("Plugin %s skipped - Failed to open plugin. Err: %v", plugPkgPath, err)
			continue
		}

		if plugSymbol, err := plugPkg.Lookup("Plug"); err == nil {
			switch valType := plugSymbol.(type) {
			case pi.ReverseProxyPlug:
				p := plugSymbol.(pi.ReverseProxyPlug)
				p.Initialize()
				reverseProxyPlugs = append(reverseProxyPlugs, p)
				pi.Log.Infof("Plug %s (%s) was succesfully loaded", p.PlugName(), p.PlugVersion())
				ret++
			default:
				pi.Log.Infof("Plugin %s skipped - Plug symbol is of ilegal type %T,  %v", plugPkgPath, plugSymbol, valType)
			}

		} else {
			pi.Log.Infof("Cant find Plug symbol in plugin: %s: %v", plugPkgPath, err)
			continue
		}
	}

	return
}

func handleRequest(h http.Handler, p pi.ReverseProxyPlug) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				pi.Log.Infof("Recovered from panic during handleRequest!\n")
				//pi.Log.Warnf("Recovered from panic during handleRequest!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
			}
		}()
		start := time.Now()
		pi.Log.Debugf("Plug %s %s RequestHook", p.PlugName(), p.PlugVersion())
		e := p.RequestHook(w, r)
		elapsed := time.Since(start)
		pi.Log.Debugf("Request-Plug %s took %s", p.PlugName(), elapsed.String())
		if e == nil {
			h.ServeHTTP(w, r)
		} else {
			pi.Log.Infof("Request-Plug returned an error %v", e)
			w.WriteHeader(http.StatusForbidden)
		}
	})
}

func HandleRequestPlugs(h http.Handler) http.Handler {
	for _, p := range reverseProxyPlugs {
		h = handleRequest(h, p)
	}
	return h
}

func HandleResponsePlugs(resp *http.Response) (e error) {
	e = nil
	defer func() {
		if r := recover(); r != nil {
			e = errors.New("plug paniced")
			pi.Log.Warnf("Recovered from panic during HandleResponsePlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}

	}()
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		pi.Log.Debugf("Plug ResponseHook: %v", p.PlugName())
		e = p.ResponseHook(resp)
		elapsed := time.Since(start)
		pi.Log.Debugf("Response-Plug %s took %s", p.PlugName(), elapsed.String())
		if e != nil {
			pi.Log.Infof("Response-Plug returned an error %v", e)
			break
		}
	}
	return
}

func HandleErrorPlugs(w http.ResponseWriter, r *http.Request, e error) {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during HandleErrorPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	pi.Log.Infof("Error-Plug received an error %v", e)
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		p.ErrorHook(w, r, e)
		elapsed := time.Since(start)
		pi.Log.Infof("Error-Plug %s took %s", p.PlugName(), elapsed.String())
	}
	w.WriteHeader(http.StatusForbidden)
}

func UnloadPlugs() {
	defer func() {
		if r := recover(); r != nil {
			pi.Log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		initialize()
	}()
	for _, p := range reverseProxyPlugs {
		p.Shutdown()
	}
}
