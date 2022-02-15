package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/IBM/go-security-plugs/rtplugs"
	"go.uber.org/zap"
)

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

	// Hook using RoundTripper
	os.Setenv("SERVING_NAMESPACE", "default")
	os.Setenv("SERVING_SERVICE", "myserver")
	rt := rtplugs.New(log)
	if rt != nil {
		defer rt.Close()
		proxy.Transport = rt.Transport(proxy.Transport)
	}

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
