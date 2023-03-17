package engine

type RoutingLayer struct {
	*Reverse
	*Crypt
}

type RoutingHeader struct {
	Layers [3]RoutingLayer
}
