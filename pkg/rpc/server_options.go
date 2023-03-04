package rpc

type ServerOptions struct {
	store     Store
	unixPath  string
	tunEnable bool
	tunPort   uint16
	tunPeers  []string
}

func (s *ServerOptions) GetTunPort() uint16 { return s.tunPort }

type ServerOption interface {
	apply(*ServerOptions)
}

type funcServerOption struct {
	f func(*ServerOptions)
}

func (fdo *funcServerOption) apply(do *ServerOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*ServerOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithDisableTunnel() ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.tunEnable = false
	})
}

func WithStore(store Store) ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.store = store
	})
}

func WithUnixPath(path string) ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.unixPath = path
	})
}

func WithTunOptions(port uint16, peers []string) ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.tunEnable = true
		o.tunPort = port
		o.tunPeers = peers
	})
}
