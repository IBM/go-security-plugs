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

When the caller manages the list of plugs (e.g. using its own config files)
use NewPlugs instead of New.
The plulist parameter would be used instead of the env RTPLUGS 
```
rt := rtplugs.NewPlugs(pluglist, log)  
if rt != nil {  
    // We have at least one activated plug
    defer rt.Close()
    reverseproxy.Transport = rt.Transport(reverseproxy.Transport)
}
```  

Use `rt.Close()` to gracefully shutdown the work of plugs

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





## Optional use of dynamic loading

!!!OPTIONAL!!!
When using dynamic loading, 
the list of SO files to dynamically load is taken from the `RTPLUGS_SO` environment variable.
Use a comma seperated list for loading more than one plug.




