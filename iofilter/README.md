# iofilter
Enables adding a filter to any io.ReadCloser.

Given an exiting 'provider' that publishes an io.ReadCloser interface 

A caller may call:
* `provider.Read(p []byte) (n int, err error)` and
* `provider.Close() error` 
use:

To examin and filter the data transfered use: 

```
  newProvider = iofilter.New(provider, filter)
```

The newProvider offers an io.ReadCloser interface.
The data is sent to filter before it is provided to the newProvider.

If the filter returns an error or panics, the data is discarded and no more data is transfered.
