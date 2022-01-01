package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/IBM/go-security-plugs/historical/reverseproxyplugs"

	"github.com/IBM/go-security-plugs/pluginterfaces"
	"go.uber.org/zap"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	ReverseProxyPlugins []string `split_words:"true"`
}

// Eample of a Reverse Proxy using plugs
func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	url, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	var h http.Handler = proxy

	var env config

	//Example env
	os.Setenv("REVERSE_PROXY_PLUGINS", "plugs/examplegate/examplegate.so")

	if err := envconfig.Process("", &env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	log.Infof("Proxy started with %v", env)

	// Have the plugins use the same logger we do
	pluginterfaces.Log = log

	// Hook the request, response and error
	if len(env.ReverseProxyPlugins) > 0 {
		reverseproxyplugs.LoadPlugs(env.ReverseProxyPlugins)
		defer reverseproxyplugs.UnloadPlugs()
		proxy.ModifyResponse = reverseproxyplugs.HandleResponsePlugs
		proxy.ErrorHandler = reverseproxyplugs.HandleErrorPlugs
		h = reverseproxyplugs.HandleRequestPlugs(proxy)
	}
	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
