# pkg/engine/transport

This is an implementation of typical network handling features, a listener,
which has an `Accept` method that returns a channel that will pick up a new
inbound connection.

## Warning

`pstoreds` and `pstoremem` both store the `libp2p.Host`'s private key in
cleartext. Consequently it is necessary to ensure to use `options.Default()` and
use an encryption key with it. The key has to be kept hot for ads, for finding
the LN node being controlled by Indra, and sending/receiving payments.

todo: need a key change protocol for this identity key that handles session
migration correctly.

## License Notes

`pstoreds` and `pstoremem` are under MIT license as seen at
the [libp2p repository](https://github.com/libp2p/go-libp2p/LICENSE).
