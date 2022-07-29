# rtplugs

The rtplugs package instruments golang http clients that supports a RoundTripper interface.
It was built and tested against [ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)

To extend reverseproxy use:
```
rt := rtplugs.New(log)  
if rt != nil {  
    // We have at least one activated plug
    defer rt.Close()
    reverseproxy.Transport = rt.Transport(reverseproxy.Transport)
}
```  
The names of plugs that will be activated is taken from the `RTPLUGS` environment variable. 
Use a comma seperated list for activating more than one plug.

The namespace used is taken from the `NAMESPACE` environment variable or the `/etc/podinfo/namespace` file (via downwards api). 

The servicename used is taken from the `SERVICENAME` environment variable or the `/etc/podinfo/servicename` file (via downwards api).

When the caller manages the list of plugs and their configuration (e.g. using its own config files)
use NewConfigrablePlugs() instead of New().
When using NewConfigrablePlugs, the caller specify the name and namespace of the service, and also the plug list and its configuration
```
rt := rtplugs.NewConfigrablePlugs(servicename, namespace, pluglist, config, log)  
if rt != nil {  
    // We have at least one activated plug
    defer rt.Close()
    reverseproxy.Transport = rt.Transport(reverseproxy.Transport)
}
```  

Use `rt.Close()` to gracefully shutdown the work of plugs. 
Graceful shutdown ensure no loss of data in plugs. 


## Optional alignment of logging facility and orderly shutdown

`log` is an optional (yet recommended in production) method to set the logger that will be used by rtplugs and all plugs. 
If `log` is set to nil, "go.uber.org/zap" Development template is used.
If `log` is set to any other logger meeting the `pluginterfaces.Logger` interface this logger will be used instead.

For example:
```diff
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

+	// Hook using RoundTripper
+	rt := rtplugs.New(log)
+	if rt != nil {
+		// We have at least one activated plug
+		defer rt.Close()
+		proxy.Transport = rt.Transport(proxy.Transport)
+	}

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
```  





