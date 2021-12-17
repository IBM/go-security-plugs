package reverseproxyplugs

import (
	goLog "log"
	"net/http"
	"plugin"
	"time"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

var reverseProxyPlugs []pluginterfaces.ReverseProxyPlug
var reverseProxyPlugNames []string

var log pluginterfaces.Logger
var plugins []string
var config map[string]interface{}

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

var defaultLog dLog

func init() {
	initialize()
}

func initialize() {
	reverseProxyPlugs = []pluginterfaces.ReverseProxyPlug{}
	reverseProxyPlugNames = []string{}
	log = defaultLog
	plugins = []string{}
	config = make(map[string]interface{})
}

func LoadPlugs(l pluginterfaces.Logger, c map[string]interface{}) (ret int) {
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("Recovered from panic during LoadPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	ret = len(reverseProxyPlugs)

	if l != nil {
		log = l
	}
	log.Infof("LoadPlugs started\n")

	if config = c; c == nil {
		log.Infof("Plugs disabled - config is empty\n")
		return
	}
	val, ok := config["reverseproxyplugins"]
	if !ok {
		log.Infof("Plugs disabled - config has no reverseproxyplugins key\n")
		return
	}

	switch valType := val.(type) {
	default:
		log.Infof("Plugs disabled - config[\"reverseproxyplugins\"] is of ilegal type %v", valType)
		return
	case []string:
		plugins = val.([]string)
	}
	for _, ext := range plugins {
		p, err := plugin.Open(ext)
		if err != nil {
			log.Infof("Plugin %s skipped - Failed to open plugin. Err: %v", ext, err)
			continue
		}

		if f, err := p.Lookup("Plug"); err == nil {
			switch valType := f.(type) {
			case pluginterfaces.ReverseProxyPlug:
				p := f.(pluginterfaces.ReverseProxyPlug)
				p.Initialize(log, config)
				reverseProxyPlugs = append(reverseProxyPlugs, p)
				reverseProxyPlugNames = append(reverseProxyPlugNames, p.PlugName())
				ret++
			default:
				log.Infof("Plugin %s skipped - Plug symbol is of ilegal type %T,  %v", ext, f, valType)
			}

		} else {
			log.Infof("Cant find Plug symbol in plugin: %s: %v", ext, err)
			continue
		}
	}

	log.Infof("Plugs %v\n", reverseProxyPlugNames)
	return
}

func handleRequest(h http.Handler, p pluginterfaces.ReverseProxyPlug) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Warnf("Recovered from panic during handleRequest!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
			}
		}()
		start := time.Now()
		log.Debugf("Starting Request-Plug %s", p.PlugName())
		log.Debugf("Plug RequestHook: %v %v: %v", p.PlugName(), p.PlugVersion(), p.PlugLogger())
		e := p.RequestHook(w, r)
		elapsed := time.Since(start)
		log.Debugf("Request-Plug %s took %s", p.PlugName(), elapsed.String())
		if e == nil {
			h.ServeHTTP(w, r)
		} else {
			log.Infof("Request-Plug returned an error %v", e)
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
			log.Warnf("Recovered from panic during HandleResponsePlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	for i, p := range reverseProxyPlugs {
		start := time.Now()
		log.Debugf("Plug ResponseHook: %v", reverseProxyPlugNames[i])
		e = p.ResponseHook(resp)
		elapsed := time.Since(start)
		log.Debugf("Response-Plug %s took %s", reverseProxyPlugNames[i], elapsed.String())
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
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		p.ErrorHook(w, r, e)
		elapsed := time.Since(start)
		log.Infof("Error-Plug %s took %s", p.PlugName(), elapsed.String())
	}
}

func UnloadPlugs() {
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("Recovered from panic during ShutdownPlugs!\n\tOne or more plugs may be skipped\n\tRecover: %v", r)
		}
	}()
	for _, p := range reverseProxyPlugs {
		p.Shutdown()
	}
	initialize()
}
