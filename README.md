# go-security-plugs
Plugs4Security

The *reverseproxyplugs* package enable extending the go reverseproxy using security enhncements by third parties introduced as shared libraries. 

## reverseproxyplugs
The package *reverseproxyplugs* extends the go reverse proxy with security plugs. 
See the *proxy* for an example how to use *reverseproxyplugs*. 
Once *reverseproxyplugs* where embedded in a go application using the go reverseproxy, it offers a unified and simple interface for extending the go application with third party secuity enhancements. 
*examplegate* is an example secuity enhancements which simply logs all request and response headers. 


To run the proxy:

1. build the examplegate using:
    cd examplegate
    go build -buildmode=plugin  .
    cd ..

2. run a sample server:
    cd server
    go run .

3. in a different window, run a sample proxy:
    cd proxy
    go run .


4. using a browser or curl try the url: http://127.0.0.1:8081   

    

