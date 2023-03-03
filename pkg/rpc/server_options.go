package rpc

type serverOptions struct {
	disableTunnel bool
	store         Store
}

type ServerOption interface {
	apply(*serverOptions)
}

type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(do *serverOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithDisableTunnel() ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.disableTunnel = true
	})
}

func WithStore(store Store) ServerOption {
	return newFuncServerOption(func(o *serverOptions) {
		o.store = store
	})
}
