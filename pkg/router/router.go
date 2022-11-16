package router

import (
	"net"
)

type Router struct {
	net.IP
}

func NewRouter(ip net.IP) *Router {
	return &Router{IP: ip}
}
