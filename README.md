# go-security-plugs
Plugs4Security

The [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) package enable extending a go application using [the standart go reverseproxy](https://go.dev/src/net/http/httputil/reverseproxy.go) with one or more *secuity extensions*.

The *security extensions* can then be introduced by third parties as shared libraries and developed seperatly from the go application. 


![image](https://github.com/IBM/go-security-plugs/blob/main/rpplugs.png)


## reverseproxyplugs

Both the extended go application and the *secuity extension* import the [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) package to gain access to the interfaces shared between the two and some utility code to help glue the application with the *secuity extensions*.

[**proxy**](https://github.com/IBM/go-security-plugs/tree/main/proxy) is an example of a go application that uses the go reverseproxy and enable a unified and simple interface for security extensions by importing and using [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs).

[**examplegate**](https://github.com/IBM/go-security-plugs/tree/main/examplegate) is an example third party secuity enhancement which import and use [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) to be able to hook go application that expose this interface.


To run the proxy:

1. build the examplegate using:
```
    cd examplegate
    go build -buildmode=plugin  .
    cd ..
```
2. run a sample server:
```
    cd server
    go run .
```
3. in a different window, run a sample proxy:
```
    cd proxy
    go run .
```

4. using a browser or curl try the url: http://127.0.0.1:8081   

    

