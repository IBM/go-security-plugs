package reverseproxyplugs

import (
	"io/ioutil"
	goLog "log"
	"net/http"
	"path/filepath"
	"plugin"
	"time"

	"github.com/IBM/go-security-plugs/pluginterfaces"
)

var reverseProxyPlugs []pluginterfaces.ReverseProxyPlug
var reverseProxyPlugNames []string
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

var defaultLog dLog

func LoadPlugs(l pluginterfaces.Logger, plugDir string, config map[string]interface{}) int {
	var extensions []string

	if log = l; log == nil {
		log = defaultLog
	}

	if plugDir == "" {

		log.Infof("Plugs disabled - no plug directory provided\n")
		return 0
	}

	dirs, err := ioutil.ReadDir(plugDir)
	if err != nil {
		log.Infof("Failed too open dir %s: %v", plugDir, err)
		return 0
	}

	for _, dirInfo := range dirs {
		if dirInfo.IsDir() {
			log.Infof("Found a plug directory: %s\n", dirInfo.Name())
			path := filepath.Join(plugDir, dirInfo.Name(), dirInfo.Name()+".so")
			extensions = append(extensions, path)
		}
	}

	for _, ext := range extensions {
		p, err := plugin.Open(ext)
		if err != nil {
			log.Infof("Failed to open plugin: %s, %v", ext, err)
			continue
		}

		if f, err := p.Lookup("Plug"); err == nil {
			p := f.(pluginterfaces.ReverseProxyPlug)
			p.Initialize(log)
			reverseProxyPlugs = append(reverseProxyPlugs, p)
			reverseProxyPlugNames = append(reverseProxyPlugNames, p.PlugName())
		} else {
			log.Infof("Cant find Plug function in plugin: %s: %v", ext, err)
			continue
		}
	}

	log.Infof("Plugs %v\n", reverseProxyPlugNames)
	return len(reverseProxyPlugs)
}

func handleRequest(h http.Handler, p pluginterfaces.ReverseProxyPlug) http.Handler {
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
