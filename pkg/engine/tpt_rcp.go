package engine

import (
	"net"
	"net/netip"
	"sync"
	
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type RCPListener struct {
	sync.Mutex
	C   ByteChan
	mtu int
	net.PacketConn
}

func NewRCPListener(addr string, mtu, bufs int) (c *RCPListener, e error) {
	var ap netip.AddrPort
	ap, e = netip.ParseAddrPort(addr)
	ipAddr := net.ParseIP(ap.Addr().String())
	nw := "udp"
	if ipAddr.To4() != nil {
		nw = "udp4"
	}
	var lis net.PacketConn
	lis, e = net.ListenPacket(nw, addr)
	c = &RCPListener{
		C:          make(ByteChan, bufs),
		mtu:        mtu,
		PacketConn: lis,
	}
	go func() {
		for {
			buf := slice.NewBytes(c.mtu)
			var n int
			var a net.Addr
			if n, a, e = c.PacketConn.ReadFrom(buf); fails(e) {
				break
			}
			go c.handle(a, buf[:n])
		}
	}()
	return
}

func (c *RCPListener) handle(addr net.Addr, b slice.Bytes) {
	c.C <- b
}
