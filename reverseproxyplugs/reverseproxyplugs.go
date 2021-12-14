package reverseproxyplugs

import (
	"net/http"
	"plugin"
	"time"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type reverseProxyPlug interface {
	Initialize(Logger)
	RequestHook(http.ResponseWriter, *http.Request) error
	ResponseHook(*http.Response) error
	ErrorHook(http.ResponseWriter, *http.Request, error)
	Shutdown()
	PlugName() string
}

var reverseProxyPlugs []reverseProxyPlug
var log Logger

func LoadPlugs(l Logger, extensions []string) {
	log = l
	for _, ext := range extensions {
		p, err := plugin.Open(ext)
		if err != nil {
			log.Infof("Failed to open plugin: %s", ext)
			continue
		}

		if f, err := p.Lookup("Plug"); err == nil {
			p := f.(reverseProxyPlug)
			p.Initialize(log)
			reverseProxyPlugs = append(reverseProxyPlugs, p)
		} else {
			log.Infof("Cant find Plug function in plugin: %s", ext)
			continue
		}

	}
	log.Infof("RPPlugs %v\n", reverseProxyPlugs)
}

func handleRequest(h http.Handler, p reverseProxyPlug) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if p.RequestHook(w, r) == nil {
			elapsed := time.Since(start)
			log.Infof("Request-Plug %s took %s", p.PlugName(), elapsed.String())
			h.ServeHTTP(w, r)
		}
	})
}

func HandleRequestPlugs(h http.Handler) http.Handler {
	for _, p := range reverseProxyPlugs {
		h = handleRequest(h, p)
	}
	return h
}

func HandleResponsePlugs(resp *http.Response) error {
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		p.ResponseHook(resp)
		elapsed := time.Since(start)
		log.Infof("Response-Plug %s took %s", p.PlugName(), elapsed.String())
	}
	return nil
}

func HandleErrorPlugs(w http.ResponseWriter, r *http.Request, e error) {
	for _, p := range reverseProxyPlugs {
		start := time.Now()
		p.ErrorHook(w, r, e)
		elapsed := time.Since(start)
		log.Infof("Error-Plug %s took %s", p.PlugName(), elapsed.String())
	}
}

func ShutdownPlugs() {
	for _, p := range reverseProxyPlugs {
		p.Shutdown()
	}
}
