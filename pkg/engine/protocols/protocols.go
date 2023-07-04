package protocols

type NetworkProtocols byte

const (
	IP4 NetworkProtocols = 1
	IP6 NetworkProtocols = 1 << iota
	// add more here if any such thing of this kind can be used ???
)
