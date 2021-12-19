# go-security-plugs
Plugs4Security

The [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) package enable safely extending a go application using [the standard go reverseproxy](https://go.dev/src/net/http/httputil/reverseproxy.go) with one or more *secuity extensions*. The task of this packag eis to not only load extensions, but also recover from any panic situations and handle all errors that can be thrown upon it from extensions. It is meant to keep the go application safe from harm done by extensions to a certain degree. It does not protect the application from extensions which: use excasive memory, cpu or other system resources (file descriptors etc.). 

*Security extensions* can then be introduced by third parties as shared libraries and developed seperatly from the go application. 

This allows loading extensions as the go application starts based on configuration as examplified in the example [**proxy**](https://github.com/IBM/go-security-plugs/tree/proxy.go)


![image](https://github.com/IBM/go-security-plugs/blob/main/security-plugs.png)


## pluginterfaces

Both the the [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) and the *secuity extension* import the [**pluginterfaces**](https://github.com/IBM/go-security-plugs/tree/main/pluginterfaces) package to gain access to the interfaces shared between the two.

## reverseproxyplugs

An application looking to extend reverseproxy use the [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) to load and communicate with the *secuity extensions*.

[**proxy**](https://github.com/IBM/go-security-plugs/tree/proxy.go) is an example of a go application that uses the go reverseproxy and enable a unified and simple interface for security extensions by importing and using [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs).

[**examplegate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/examplegate) is an example third party secuity enhancement which import and use [**pluginterfaces**](https://github.com/IBM/go-security-plugs/tree/main/pluginterfaces) and can be loaded by the [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs).

When using go plugs one must ensure that the shared library uses the same package versions as the application. To ensure all plugs use the same package versions as your main app:
1. Clone the plugs into the plugs directory of yout app (as shown here with the [**examplegate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/examplegate) plug).
2. Build all plugs in your plugs/ directory (See [**buildPlugs.sh**](https://github.com/IBM/go-security-plugs/blob/main/buildPlugs.sh)) before building/running your app. 




To run the example here:

1. build and run a sample http server:
```
    ./runServer.sh
```
2. in a different window, run build the Plugs and run the proxy:
```
    ./runProxyWithPlugs.sh
```

3. using a browser or curl try the url: http://127.0.0.1:8081   and see the logs pile up in the proxy window.

   


    

