// This is a deprecated version
//
// Use instead github.com/IBM/go-security-plugs/pluginterfaces
package pluginterfaces

import (
	"net/http"

	"go.uber.org/zap"
)

// This is a deprecated version
//
// Use instead github.com/IBM/go-security-plugs/pluginterfaces
//
// Any logger of this interface can be used by rtlog and all connected plgins
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Sync() error
}

// This is a deprecated version
//
// Use instead github.com/IBM/go-security-plugs/pluginterfaces
//
// The logger for the rtplugs and all connected plgins
var Log Logger

// This is a deprecated version
//
// Use instead github.com/IBM/go-security-plugs/pluginterfaces
//
// A plugin based on the older ReverseProxyPlug
type ReverseProxyPlug interface {
	Initialize()
	Shutdown()
	PlugName() string
	PlugVersion() string
	RequestHook(http.ResponseWriter, *http.Request) error
	ResponseHook(*http.Response) error
	ErrorHook(http.ResponseWriter, *http.Request, error)
}

func init() {
	logger, _ := zap.NewDevelopment()
	Log = logger.Sugar()
}
