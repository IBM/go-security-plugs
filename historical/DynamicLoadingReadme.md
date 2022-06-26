# go-security-plugs
Plugs4Security


The [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) package uses the http RoundTripper interfce to enable safely extending any go application that uses the http client. More specifically it is designed to enable extensing the [standard go reverseproxy](https://go.dev/src/net/http/httputil/reverseproxy.go) with one or more *secuity extensions*. 

The package not only load extensions, but also recover from any panic situations and handle all errors that may be introduced by extensions. It is meant to keep the go application safe from harm done by extensions to a certain degree. It does not protect the application from extensions which: use excasive memory, cpu or other system resources (file descriptors etc.). 

Using [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs), *secuity extensions* of a reverseproxy may:

1. ___Block the request___ before it reaches the server. Blocking the reqeust will result in the connection to the client being closed.  The client will receive a 502 response code. The request will never reach the server.

2. ___Block the response___ from the server before it is returned to the client. Blocking the response will result in the connection to the client being closed. The client will receive a 502 response code and no data will be transfered from the server to the client. The connection to the server will also be closed, signaling to the server that the client disconnected and no further service is required. 

3. ___Asynchroniously cancel a request___ while it is being processed by the server. Canceling the request will result in the connection to the client and server being closed. No additional data (beyond what was already delivered prior to request cancelation) will be further delivered from the server to the client. There are two cases to consider:

    1. The request was cancled __before__ the response code was sent to the client. In this case, the client will now receive a 502 response code.  Closing the connection to the server will signal to the server that the client disconnected and no further service is required.

    2.  The request was cancled __after__ the response code was sent to the client. In this case, closing the connection to the client will signal to the client that the server aborted the service. Closing the connection to the server will signal to the server that the client disconnected and no further service is required. 


*Security extensions* such as [**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) can be introduced by third parties as packages and developed seperatly from the main application. Such extensions can later be pluged using [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) to the application. As demonstrated by [**proxy**](https://github.com/IBM/go-security-plugs/tree/proxy.go), [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) supports both static and dynamic loading the extensions on-demand, based, on the application configuration.

<p align="center">
    <img src="https://github.com/IBM/go-security-plugs/blob/main/rtplugins.png" width="700"  />
</p>


## pluginterfaces

Both the the  [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) and the *secuity extension* import the [**pluginterfaces**](https://github.com/IBM/go-security-plugs/tree/main/pluginterfaces) package to gain access to the interfaces shared between the two.


## rtplugs

An application looking to extend reverseproxy (or any other http client) use the [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs) to load and communicate with the *secuity extensions*.

[**proxy**](https://github.com/IBM/go-security-plugs/tree/proxy.go) is an example of a go application that uses the go reverseproxy and enable a unified and simple interface for security extensions by importing and using [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs).

## rtgate

[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) is an example third party secuity enhancement which import and use [**pluginterfaces**](https://github.com/IBM/go-security-plugs/tree/main/pluginterfaces) and can be loaded by the [**rtplugs**](https://github.com/IBM/go-security-plugs/tree/main/rtplugs).

[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) demonstrates how a request can be canceled before reaching the server. It will block any request that include the header "X-Block-Req:true". 


[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) demonstrates how a response can be canceled before reaching the client. It will block the response of any request that include the header "X-Block-Resp:true". 


[**rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate) demonstrates how a request can be canceled asynchrniously using a security extension. The code allows requests to last for no more than 5 seconds by default. Alternativly timeout can be specified using the reqeust header "X-Block-Async:<duration>". For example: "X-Block-Async:3s" results in a cancel being processed 3 seconds from request. The timeout examplifies an asynchrnious decission to  cancel a request after it was delivered for processing by the server. 

# How to use - static imports of plugs
Static loading does not require the use of CGO and is exmaplified by the runProxyStatic.sh script
 
Run the `runProxyStatic.sh` script, 

* `RTPLUGS_PKG` defines the list of plugs you wish to import
* `RTPLUGS` defines the list of plugs you wish to activate

# How to use - dynamic imports of plugs
Dynamic loading requires the use of CGO and is exmaplified by the runProxyDyn.sh script

Run the `runProxyDyn.sh` script, Make sure to set the list of plugs you wish to import and activate

* `RTPLUGS_PKG` defines the list of plugs you wish to import
* `RTPLUGS` defines the list of plugs you wish to activate

To run the example here:

1. build and run a sample http server:
```
    ./runServer.sh
```
2. in a different window, run build the Plugs and run the proxy:
```
    ./runProxyStatic.sh 
```
or 
```
    ./runProxyDyn.sh 
```

3. using a browser or curl try the url: http://127.0.0.1:8081   and see the logs pile up in the proxy window.
   
    Here are some examples: 
```
    curl 127.0.0.1:8081 -v   
    curl 127.0.0.1:8081 -v   -H "X-Block-Req:true"
    curl 127.0.0.1:8081 -v   -H "X-Block-Resp:true"
    curl 127.0.0.1:8081 -v   -H "X-Sleep:4s"
    curl 127.0.0.1:8081 -v   -H "X-Sleep:4s"   -H "X-Block-Async:2s"
    curl 127.0.0.1:8081 -v   -H "X-Sleep:0.1s" -H "X-Sleep-Step:0.001s" -H "X-Sleep-Num-Steps:10000"
    curl 127.0.0.1:8081 -v   -H "X-Sleep:0.1s" -H "X-Sleep-Step:0.001s" -H "X-Sleep-Num-Steps:10000"  -H "X-Block-Async:2s"
```

The reqeust headers "X-Sleep", "X-Sleep-Step", "X-Sleep-Num-Steps" control the behaviour of our  [**sample server**](https://github.com/IBM/go-security-plugs/tree/main/server). 
    

The reqeust headers  "X-Block-Req", "X-Block-Resp",  "X-Block-Async" control the behaviour of our  [**sample rtgate**](https://github.com/IBM/go-security-plugs/tree/main/plugs/rtgate). 

# Historical Notes
See [**historical**](https://github.com/IBM/go-security-plugs/tree/main/historical) for alternative hooks used in previous versions of this code.
   


    

