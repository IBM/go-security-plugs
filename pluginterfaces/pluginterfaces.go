package pluginterfaces

import "net/http"

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type ReverseProxyPlug interface {
	Initialize(Logger)
	RequestHook(http.ResponseWriter, *http.Request) error
	ResponseHook(*http.Response) error
	ErrorHook(http.ResponseWriter, *http.Request, error)
	Shutdown()
	PlugName() string
}
