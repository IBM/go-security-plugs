package reverseproxyplugs

import (
	"fmt"
	"io/ioutil"
	goLog "log"
	"net/http"
	"path/filepath"
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
var reverseProxyPlugNames []string
var log Logger

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

func LoadPlugs(l Logger, extensions []string) int {
	if log == nil {
		log = defaultLog
	} else {
		log = l
	}
	dirs, err := ioutil.ReadDir("plugs")
	if err != nil {
		panic(err)
	}

	for _, dirInfo := range dirs {
		if dirInfo.IsDir() {
			fmt.Println("Found plug in plugs dir: ", dirInfo.Name())
			path := filepath.Join("plugs", dirInfo.Name(), dirInfo.Name()+".so")
			extensions = append(extensions, path)
		}
	}
	/*
		files, _ := ioutil.ReadDir("plugs")
		for _, file := range files {
			if file.Name() != "extpoints.go" && !strings.HasSuffix(file.Name(), "_ext.go") {
				path := filepath.Join(packagePath, file.Name())
				log.Printf("Processing file %s", path)
				packageName, ifaces = processFile(path)
				if len(ifacesAllowed) > 0 {
					var ifacesFiltered []string
					for _, iface := range ifaces {
						_, allowed := ifacesAllowed[iface]
						if allowed {
							ifacesFiltered = append(ifacesFiltered, iface)
						}
					}
					ifaces = ifacesFiltered
				}
				log.Printf("Found interfaces: %#v", ifaces)
				ifacesAll = append(ifacesAll, ifaces...)
			}
		}
	*/
	for _, ext := range extensions {
		p, err := plugin.Open(ext)
		if err != nil {
			log.Infof("Failed to open plugin: %s, ", ext, err)
			continue
		}

		if f, err := p.Lookup("Plug"); err == nil {
			p := f.(reverseProxyPlug)
			p.Initialize(log)
			reverseProxyPlugs = append(reverseProxyPlugs, p)
			reverseProxyPlugNames = append(reverseProxyPlugNames, p.PlugName())
		} else {
			log.Infof("Cant find Plug function in plugin: %s", ext)
			continue
		}

	}
	log.Infof("Plugs %v\n", reverseProxyPlugNames)
	return len(reverseProxyPlugs)
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
