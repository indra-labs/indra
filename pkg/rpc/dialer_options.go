package rpc

// dialOptions configure a Dial call. dialOptions are set by the DialOption
// values passed to Dial.
type dialOptions struct {
	key               RPCPrivateKey
	peerPubKey        RPCPublicKey
	peerRPCIP         string
	keepAliveInterval int
}

// DialOption configures how we set up the connection.
type DialOption interface {
	apply(*dialOptions)
}

// funcDialOption wraps a function that modifies dialOptions into an
// implementation of the DialOption interface.
type funcDialOption struct {
	f func(*dialOptions)
}

func (fdo *funcDialOption) apply(do *dialOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
	return &funcDialOption{
		f: f,
	}
}

type joinDialOption struct {
	opts []DialOption
}

func (jdo *joinDialOption) apply(do *dialOptions) {
	for _, opt := range jdo.opts {
		opt.apply(do)
	}
}

func newJoinDialOption(opts ...DialOption) DialOption {
	return &joinDialOption{opts: opts}
}

func WithKeepAliveInterval(seconds int) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.keepAliveInterval = seconds
	})
}

func WithPeer(pubKey string) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.peerPubKey = DecodePublicKey(pubKey)
	})
}

func WithPrivateKey(key string) DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.key = DecodePrivateKey(key)
	})
}
