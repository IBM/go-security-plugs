
This is a deprecated version

Use instead github.com/IBM/go-security-plugs/rtplugs

# Historical Notes
Prior to the use of the RoundTripper interface, the implementation used the proxy's: `proxy.ModifyResponse`, `proxy.ErrorHandler`, and a wrapper to the `http.Handler`. This option was abandoned and the code was left here, either commented out (in examples), or at the no-longer-used [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/historical/reverseproxyplugs)  package.
    
## reverseproxyplugs

An older [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/historical/reverseproxyplugs) that does not use the RoundTripper interface is also included in this repository. 

<p align="center">
    <img src="https://github.com/IBM/go-security-plugs/blob/main/historical/reverseproxyplugs.png" width="700"  />
</p>

The [**reverseproxyplugs**](https://github.com/IBM/go-security-plugs/tree/main/reverseproxyplugs) has no apperent advetages over the new mechanism that uses the RoundTripper and is left here fully functional for reference and as a second option. The use of this option is included in [**historical/proxy**](https://github.com/IBM/go-security-plugs/tree/historical/proxy.go) code. 

   


    

