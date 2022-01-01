# RtGate

This plug include an example RoundTripper gate that logs all request and response headers

The plug also timeout any request asynchrniously after 5 seconds. This timeout examplifies the ability to asynchrniously cancel a request mid-way while it is being processsed if and when a security gate determines that the reqeust should be terminated.

