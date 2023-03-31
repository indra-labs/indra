package engine

import (
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/packet"
	"git-indra.lan/indra-labs/indra/pkg/rudp"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type RUDPBuffer map[pub.Bytes][]slice.Bytes

type RUDP struct {
	endpoint      *netip.AddrPort
	in            ByteChan
	conn          *rudp.Conn
	good, corrupt atomic.Uint64
	failRate      atomic.Uint32
	bufMx         sync.Mutex
	buffers       RUDPBuffer
}

func NewRUDP(remote *netip.AddrPort, local net.IP, bufs,
	mtu int) (r *RUDP, e error) {
	
	if mtu <= 0 {
		// Conservative packet size limit.
		mtu = 1382
	}
	raddr := net.UDPAddrFromAddrPort(*remote)
	laddr := &net.UDPAddr{IP: local}
	var conn *net.UDPConn
	if conn, e = net.DialUDP("udp", laddr, raddr); check(e) {
		return
	}
	rUDPConn := rudp.NewConn(conn, rudp.New())
	r = &RUDP{
		endpoint: remote,
		in:       make(ByteChan, bufs),
		conn:     rUDPConn,
		buffers:  make(RUDPBuffer),
	}
	r.good.Store(256)
	r.corrupt.Store(1)
	listener := rudp.NewListener(conn)
	buf := make(slice.Bytes, mtu)
	quit := qu.T()
	go func() {
		for {
			total := r.good.Load() + r.corrupt.Load()
			// Compute average out of 256 as 100% vs 0% and average
			// with previous failRate.
			x := 256*total/(1+r.corrupt.Load()) - 1 // avoiding divide by zero
			r.failRate.Store(uint32(byte((x + uint64(r.failRate.Load())) / 2)))
			if total > 1<<24 {
				// After 16mb of traffic we reset the counters.
				r.good.Store(0)
				r.corrupt.Store(0)
			}
			var rconn *rudp.Conn
			rconn, e = listener.AcceptRudp()
			if check(e) {
				break
			}
			var n int
			if n, e = rconn.Read(buf); check(e) {
				if e.Error() == "corrupt" {
					r.corrupt.Add(uint64(n))
					continue
				}
				continue
			}
			r.good.Add(uint64(n))
			
			b := buf[:n]
			var fromPub *pub.Key
			var toCloak cloak.PubKey
			if fromPub, toCloak, e = packet.GetKeys(b); check(e) {
				continue
			}
			_ = fromPub
			_ = toCloak
			// todo: collect and manage packet buffers.
			r.in <- buf[:n]
			
			select {
			case <-quit:
				return
			}
		}
	}()
	return
}

func (k *RUDP) Send(b slice.Bytes) (e error) {
	// todo: split into packets.
	if _, e = k.conn.Write(b); check(e) {
	}
	return
}

func (k *RUDP) Receive() <-chan slice.Bytes {
	return k.in
}
