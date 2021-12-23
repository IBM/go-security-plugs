# SampleServer

This sample server help develop and test [**go-security-plugs**](https://github.com/IBM/go-security-plugs).

Running this server can be done in this directory by:
'''
    go build .
'''

Example use for a request that will first wait 1 second, than start sending 1000 response lines, 10 ms apart: 
'''
    curl 127.0.0.1:8080 -v -H "X-Sleep:1s" -H "X-Sleep-Step:10ms" -H "X-Sleep-Num-Steps:1000"
'''