/*
Package http implements "http" probe type that test HTTP(S) servers.

The value of the probe will be 0 if HTTP server responds with status
between 200 and 299, or 1.0 for other status values, connection errors,
or timeouts.

If parse is true, the response body will be interpreted
as a floating point number, and will be used as the probe value.

Basic authentication can be used by embedding user:password in url.

The constructor takes these parameters:

    Name       Type     Default   Description
	kubeConfig string   ""        kubernetes config file path
*/
package monitorA
