# pkg/engine/transport

This is an implementation of typical network handling features, a listener,
which has an `Accept` method that returns a channel that will pick up a new
inbound connection. (todo: is there a proper interface with such a method?)

(answer to todo: plans afoot to make it a standard net.Listener with the minor
caveat that the `Addr` function only allows one address in return and we support
multiple bound addresses.)