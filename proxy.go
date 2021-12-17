package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/IBM/go-security-plugs/reverseproxyplugs"
	"go.uber.org/zap"
)

// Eample of a Reverse Proxy using plugs
func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	env := make(map[string]interface{})
	env["reverseproxyplugins"] = []string{"plugs/examplegate/examplegate.so"}

	// Load the list of shared libraries as defined by env.extensions
	// Set the shared library to use the application log facilities
	// Log facilities interface include: Debugf, Infof, Warnf, Errorf
	reverseproxyplugs.LoadPlugs(log, env)
	defer reverseproxyplugs.UnloadPlugs()

	url, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	// Hook the request, response and error
	h := reverseproxyplugs.HandleRequestPlugs(proxy)
	proxy.ModifyResponse = reverseproxyplugs.HandleResponsePlugs
	proxy.ErrorHandler = reverseproxyplugs.HandleErrorPlugs

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
