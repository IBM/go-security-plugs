package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	_ "github.com/IBM/go-security-plugs/plugs/testgate"
	"github.com/IBM/go-security-plugs/qpsecurity"
	"go.uber.org/zap"
	"knative.dev/serving/pkg/queue/sharedmain"
)

// Eample of a Reverse Proxy using plugs
func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	url, err := url.Parse("http://127.0.0.1:8889")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	var h http.Handler = proxy

	// Hook using RoundTripper
	os.Setenv("SERVING_NAMESPACE", "default")
	os.Setenv("SERVING_SERVICE", "myserver")
	os.Setenv("RTPLUGS", "testgate")
	//rt := rtplugs.New(log)
	//if rt != nil {
	//	defer rt.Close()
	//	proxy.Transport = rt.Transport(proxy.Transport)
	//}
	qp := qpsecurity.NewQPSecurityPlugs()
	d := new(sharedmain.Defaults)
	d.Ctx = context.Background()
	d.Logger = log
	d.Transport = proxy.Transport

	qp.Setup(d)
	defer qp.Shutdown()
	proxy.Transport = d.Transport
	log.Infof("Transport ready")

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
