# qu
### observable signal channels

This is a wrapper around `chan struct{}` that forgives some common mistakes 
like sending on closed channels and closing closed channels, as well as 
printing logs about when channels are created, waited on, sent to and closed.

This library makes debugging concurrent code a lot easier. IMHO. YMMV.
