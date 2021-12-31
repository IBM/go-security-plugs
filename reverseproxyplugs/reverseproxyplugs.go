package reverseproxyplugs

import (
	"errors"
	goLog "log"
	"net/http"
	"plugin"
	"time"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

var reverseProxyPlugs []pluginterfaces.ReverseProxyPlug
var log pluginterfaces.Logger

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
func (dLog) Sync() error {
	return nil
}

func init() {
	initialize()
}

func initialize() {
	reverseProxyPlugs = []pluginterfaces.ReverseProxyPlug{}
	log = dLog{}
}

func LoadPlugs(l pluginterfaces.Logger, plugins []string) (ret int) {
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("Recovered from panic during LoadPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	ret = len(reverseProxyPlugs)

	if l != nil {
		log = l
	}
	for _, plugPkgPath := range plugins {
		plugPkg, err := plugin.Open(plugPkgPath)
		if err != nil {
			log.Infof("Plugin %s skipped - Failed to open plugin. Err: %v", plugPkgPath, err)
			continue
		}

		if plugSymbol, err := plugPkg.Lookup("Plug"); err == nil {
			switch valType := plugSymbol.(type) {
			case pluginterfaces.ReverseProxyPlug:
				p := plugSymbol.(pluginterfaces.ReverseProxyPlug)
				p.Initialize(log)
				reverseProxyPlugs = append(reverseProxyPlugs, p)
				log.Infof("Plug %s (%s) was succesfully loaded", p.PlugName(), p.PlugVersion())
				ret++
			default:
				log.Infof("Plugin %s skipped - Plug symbol is of ilegal type %T,  %v", plugPkgPath, plugSymbol, valType)
			}

		} else {
			log.Infof("Cant find Plug symbol in plugin: %s: %v", plugPkgPath, err)
			continue
		}
	}

	return
}

func handleRequest(h http.Handler, p pluginterfaces.ReverseProxyPlug) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Infof("Recovered from panic during handleRequest!\n")
				//log.Warnf("Recovered from panic during handleRequest!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
			}
		}()
		start := time.Now()
		log.Debugf("Plug %s %s RequestHook", p.PlugName(), p.PlugVersion())
		e := p.RequestHook(w, r)
		elapsed := time.Since(start)
		log.Debugf("Request-Plug %s took %s", p.PlugName(), elapsed.String())
		if e == nil {
			h.ServeHTTP(w, r)
		} else {
			log.Infof("Request-Plug returned an error %v", e)
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
			log.Warnf("Recovered from panic during HandleResponsePlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}

	}()
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		log.Debugf("Plug ResponseHook: %v", p.PlugName())
		e = p.ResponseHook(resp)
		elapsed := time.Since(start)
		log.Debugf("Response-Plug %s took %s", p.PlugName(), elapsed.String())
		if e != nil {
			log.Infof("Response-Plug returned an error %v", e)
			break
		}
	}
	return
}

func HandleErrorPlugs(w http.ResponseWriter, r *http.Request, e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("Recovered from panic during HandleErrorPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	log.Infof("Error-Plug received an error %v", e)
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		p.ErrorHook(w, r, e)
		elapsed := time.Since(start)
		log.Infof("Error-Plug %s took %s", p.PlugName(), elapsed.String())
	}
	w.WriteHeader(http.StatusForbidden)
}

func UnloadPlugs() {
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
		initialize()
	}()
	for _, p := range reverseProxyPlugs {
		p.Shutdown()
	}
}
