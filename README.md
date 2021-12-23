# go-security-plugs
Plugs4Security


The [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) package  uses the go client RoundTripper interfce and enable safely extending a go application that uses a standard golang http clien. It specifically enable extensing the [standard go reverseproxy](https://go.dev/src/net/http/httputil/reverseproxy.go) with one or more *secuity extensions*. 

The package not only load extensions, but also recover from any panic situations and handle all errors  from extensions. It is meant to keep the go application safe from harm done by extensions to a certain degree. It does not protect the application from extensions which: use excasive memory, cpu or other system resources (file descriptors etc.). 

Using [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs), *secuity extensions* may:

1. ___Block the request___ before it reaches the server. Blocking the reqeust will result in the connection to the client being closed.  The client will receive a 502 response code. The request will never reach the server.

2. ___Block the response___ from the server before it is returned to the client. Blocking the response will result in the connection to the client being closed. The client will receive a 502 response code and no data will be transfered from the server to the client. The connection to the server will also be closed, signaling to the server that the client disconnected and no further service is required. 

3. ___Asynchroniously cancel a request___ while it is being processed by the server. Canceling the request will result in the connection to the client and server being closed. No additional data (beyond what was already delivered prior to request cancelation) will be further delivered from the server to the client. There are two cases to consider:

    1. The request was cancled __before__ the response code was sent to the client. In this case, the client will now receive a 502 response code.  Closing the connection to the server will signal to the server that the client disconnected and no further service is required.

    2.  The request was cancled __after__ the response code was sent to the client. In this case, closing the connection to the client will signal to the client that the server aborted the service. Closing the connection to the server will signal to the server that the client disconnected and no further service is required. 


![image](https://github.com/IBM/go-security-plugs/blob/main/rtplugins.png)

*Security extensions* can then be introduced by third parties as shared libraries and developed seperatly from the go application. 

This allows loading extensions as the go application starts based on configuration as examplified in the example [**proxy**](https://github.com/IBM/go-security-plugs/tree/proxy.go)

An older [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) that does not use the RoundTripper interfce is also included in this repository. 

![image](https://github.com/IBM/go-security-plugs/blob/main/reverseproxyplugs.png)


## pluginterfaces

Both the the [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) and the *secuity extension* import the [**pluginterfaces**](https://github.com/IBM/go-security-plugs/tree/main/pluginterfaces) package to gain access to the interfaces shared between the two.


## rtplugs

An application looking to extend reverseproxy (or any other http client) use the [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) to load and communicate with the *secuity extensions*.

[**proxy**](https://github.com/IBM/go-security-plugs/tree/proxy.go) is an example of a go application that uses the go reverseproxy and enable a unified and simple interface for security extensions by importing and using [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs).

## rtgate

[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) is an example third party secuity enhancement which import and use [**pluginterfaces**](https://github.com/IBM/go-security-plugs/tree/main/pluginterfaces) and can be loaded by the [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs).

[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) demonstrates how a request can be canceled before reaching the server. It will block any request that include the header key "X-Block-Req". 


[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) demonstrates how a response can be canceled before reaching the client. It will block the response of any request that include the header key "X-Block-Resp". 


[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) demonstrates how a request can be canceled asynchrniously using a security extension. The code allows requests to last for no more than 5 seconds by default. Alternativly timeout can be specified using the reqeust header key "X-Block-Async". For example: "X-Block-Async:3s" results in a cancel being processed 3 seconds from request. The demonstrated timeout is just example for an asynchrnious decission to  cancel the request. 

# How to use

When using go plugs one must ensure that the shared library uses the same package versions as the application. To ensure all plugs use the same package versions as your main app:
1. Clone the plugs into the plugs directory of yout app (as shown here with the [**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) plug).
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

   

## reverseproxyplugs

An older version extended the reverseproxy without using the roundtripper. This [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) option has no apperent advetages over the new mechanism that uses the RoundTripper and is left here for reference only. 

To use it uncommend the proxy's commented lines that uses the [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs). 

   


    

