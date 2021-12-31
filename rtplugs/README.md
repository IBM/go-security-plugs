# rtplugs

The rtplugs package instruments golang http clients that supports a RoundTripper interface.
It was built and tested against [ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)

To extend reverseproxy use:
```
rt := rtplugs.New(pluginList)
if rt != nil {
    defer rt.Close()
    reverseproxy.Transport = rt.Transport(reverseproxy.Transport)
}
```  
While `pluginList` is a slice of strings for the path of plugins (.so files) to load

You may set the logger that will be used by rtplugs and all plugins by setting 
`pluginterfaces.Log` to any logger thet meets the `pluginterfaces.Logger` interface.

Use `rt.Close()` to gracefully shutdown the work of plugins

For example:
```
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

	// Have the plugins use the same logger we do    
    pluginterfaces.Logger = log  // (optional)

    // Hook using RoundTripper
    rt := rtplugs.New(pluginList)
    if rt != nil {
        defer rt.Close()
        proxy.Transport = rt.Transport(proxy.Transport)
    }

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
```  
