package pluginterfaces

import "net/http"

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Sync() error
}

type AnyPlug interface {
	Shutdown()
	PlugName() string
	PlugVersion() string
}

type ReverseProxyPlug interface {
	Initialize(Logger)
	AnyPlug
	RequestHook(http.ResponseWriter, *http.Request) error
	ResponseHook(*http.Response) error
	ErrorHook(http.ResponseWriter, *http.Request, error)
}

type RoundTripPlug interface {
	AnyPlug
	ApproveRequest(*http.Request) (*http.Request, error)
	ApproveResponse(*http.Request, *http.Response) (*http.Response, error)
}
