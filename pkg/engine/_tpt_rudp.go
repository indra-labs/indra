package engine

import (
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/rudp"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func LoadKeySlot(pr *prv.Key, pb *pub.Key) (k *packet.KeySlot) {
	return &packet.KeySlot{&Keys{Pub: pb, Bytes: pb.ToBytes(), Prv: pr}, time.Now()}
}

const (
	RUDPHandshakeLen = nonce.IVLen + 2*pub.KeyLen
	KeychangeMagic   = "KEYC"
)

type RUDPBuffer map[nonce.ID][]*Packet

type KeySlot struct {
	*Keys
	lastUsed time.Time
}

type KeyChain map[string]*KeySlot

type RUDP struct {
	*Keys
	recvKeys, sendKeys KeyChain
	endpoint           *netip.AddrPort
	in                 ByteChan
	conn               *rudp.Conn
	good, corrupt      atomic.Uint64
	failRate           atomic.Uint32
	bufMx              sync.Mutex
	buffers            RUDPBuffer
	quit               qu.C
}

func (r *RUDP) Mx(fn func()) {
	r.bufMx.Lock()
	fn()
	r.bufMx.Unlock()
}

func (r *RUDP) GetParity() (parity byte) {
	fr := r.failRate.Load()
	p := fr * 10 / 11 // 10% higher than fail rate.
	if p > 255 {
		p = 255
	}
	parity = byte(256 - p)
	return
}

func NewListenerRUDP(idKeys *Keys, bindAddress net.IP, bufs,
	mtu int, quit qu.C) (r *RUDP, e error) {
	
	if mtu <= 512 {
		// Conservative packet size limit.
		mtu = 1382
	}
	addr := &net.UDPAddr{IP: bindAddress, Port: 0}
	var conn *net.UDPConn
	if conn, e = net.ListenUDP("udp", addr); fails(e) {
		return
	}
	rUDPConn := rudp.NewConn(conn, rudp.New())
	r = MakeRUDP(idKeys, nil, rUDPConn, bufs, quit)
	buf := slice.NewBytes(mtu)
	go r.listen(conn, buf, quit)
	return
}

func (r *RUDP) listen(conn *net.UDPConn, buf slice.Bytes,
	quit qu.C) {
	
	var e error
	listener := rudp.NewListener(conn)
	for {
		total := r.good.Load() + r.corrupt.Load()
		// Compute average out of 256 as 100% vs 0% and average
		// with previous failRate.
		ratio := 256 * total / (1 + r.corrupt.Load()) // avoiding divide by zero
		r.failRate.Store(uint32(byte((ratio + uint64(r.failRate.Load())) / 2)))
		if total > 1<<24 {
			// After 16mb of traffic we reset the counters.
			r.good.Store(0)
			r.corrupt.Store(0)
		}
		var rConn *rudp.Conn
		if rConn, e = listener.AcceptRudp(); fails(e) {
			break
		}
		var n int
		addr := rConn.RemoteAddr().String()
		if n, e = rConn.Read(buf); fails(e) {
			if e.Error() == "corrupt" {
				r.corrupt.Add(uint64(n))
				continue
			}
			continue
		}
		r.good.Add(uint64(n))
		s := NewSpliceFrom(buf[:n])
		var magic string
		s.ReadMagic4(&magic)
		switch magic {
		case PacketMagic:
			var fromPub *pub.Key
			var toCloak cloak.PubKey
			var iv nonce.IV
			if fromPub, toCloak, iv, e = GetPacketKeys(s.GetAll()); fails(e) {
				continue
			}
			var to *prv.Key
			r.Mx(func() {
				for i := range r.recvKeys {
					if cloak.Match(toCloak, r.recvKeys[i].Bytes) {
						to = r.recvKeys[i].Prv
						return
					}
				}
			})
			var pkt *Packet
			if pkt, e = DecodePacket(s.GetAll(), fromPub, to, iv); fails(e) {
				continue
			}
			// The minimum pieces need to recover the packet are computable
			// by the length field and the length of the data field.
			// If the number in the buffer exceeds this number we want to
			// attempt reassembly.
			lq := int(pkt.Length) / len(pkt.Data)
			lm := int(pkt.Length) % len(pkt.Data)
			if lm != 0 {
				lq++
			}
			var buffers []*Packet
			r.Mx(func() {
				r.buffers[pkt.ID] = append(r.buffers[pkt.ID], pkt)
				if len(r.buffers[pkt.ID]) >= lq {
					buffers = r.buffers[pkt.ID]
				}
			})
			if buffers != nil {
				var msg []byte
				if buffers, msg, e = JoinPackets(buffers); fails(e) {
					// Pass it back afterwards as the JoinPackets function
					// cleans stuff that doesn't need to be done twice.
					r.Mx(func() { r.buffers[pkt.ID] = buffers })
					continue
				}
				// todo: handling stale stuff that never decoded.
				r.in <- msg
			}
		case KeychangeMagic:
			if s.Len() < 4+pub.KeyLen {
				log.D.F("message too short to contain a public key")
				continue
			}
			var peerKey *pub.Key
			s.ReadPubkey(&peerKey)
			r.Mx(func() {
				if _, ok := r.sendKeys[addr]; ok {
					r.sendKeys[addr] = LoadKeySlot(nil, peerKey)
				}
			})
		default:
			// If it is not a message, it is the delivery of a receiver public
			// key for the client, and we generate a new one and send it back
			// encrypted in a packet addressed to the provided key.
			if n < RUDPHandshakeLen {
				log.D.Ln("message too short for handshake")
				continue
			}
			s.SetCursor(0)
			var pK *pub.Key
			s.ReadPubkey(&pK)
			if pK == nil {
				continue
			}
			var iv nonce.IV
			s.ReadIV(&iv)
			if fails(Encipher(s.GetRest(), iv, r.Prv, pK, "handshake decode")) {
				continue
			}
			var peerKey *pub.Key
			s.ReadPubkey(&peerKey)
			r.Mx(func() { r.sendKeys[addr] = LoadKeySlot(nil, peerKey) })
			var recvPrv *prv.Key
			if recvPrv, e = prv.GenerateKey(); fails(e) {
				continue
			}
			recvPub := pub.Derive(recvPrv)
			r.Mx(func() { r.recvKeys[addr] = LoadKeySlot(recvPrv, recvPub) })
			// Encode the reply.
			o := NewSplice(pub.KeyLen + 4).Magic4(KeychangeMagic).Pubkey(recvPub)
			var pkt slice.Bytes
			if pkt, e = EncodePacket(PacketParams{
				To:     pK,
				From:   recvPrv,
				Parity: int(r.GetParity()),
				Length: o.Len(),
				Data:   o.GetAll(),
			}); fails(e) {
				continue
			}
			if n, e = r.conn.Write(pkt); fails(e) &&
				n != len(pkt) {
				continue
			}
			log.D.Ln("sent key change message reply")
		}
		select {
		case <-quit:
			if e = rConn.Close(); fails(e) {
			}
			return
		default:
		}
	}
}

func NewOutboundRUDP(idKeys *Keys, remote *netip.AddrPort,
	rKey *pub.Key, local net.IP, bufs, mtu int, quit qu.C) (r *RUDP,
	e error) {
	
	if mtu <= 512 {
		mtu = 1382 // Conservative packet size limit.
	}
	raddr, laddr := net.UDPAddrFromAddrPort(*remote), &net.UDPAddr{IP: local}
	var conn *net.UDPConn
	if conn, e = net.DialUDP("udp", laddr, raddr); fails(e) {
		return
	}
	r = MakeRUDP(idKeys, remote, rudp.NewConn(conn, rudp.New()), bufs, quit)
	var recvKey, ciphKey *Keys
	if recvKey, ciphKey, e = Generate2Keys(); fails(e) {
		return
	}
	reply, iv := NewSplice(RUDPHandshakeLen), nonce.New()
	reply.Pubkey(ciphKey.Pub).IV(iv)
	start := reply.GetCursor()
	reply.Pubkey(recvKey.Pub)
	if fails(Encipher(reply.GetFrom(start), iv, ciphKey.Prv, rKey,
		"handshake encode")) {
		return
	}
	var n int
	// Send the connection receiver public key to the other side.
	if n, e = r.conn.Write(reply.GetAll()); fails(e) &&
		n != RUDPHandshakeLen {
		return
	}
	addr := raddr.String()
	r.Mx(func() {
		r.recvKeys[addr] = LoadKeySlot(recvKey.Prv, recvKey.Pub)
		// Populate so the listener loads the reply key in here:
		r.sendKeys[addr] = LoadKeySlot(nil, nil)
	})
	buf := slice.NewBytes(mtu)
	go r.listen(conn, buf, quit)
	return
}

func MakeRUDP(idKeys *Keys, remote *netip.AddrPort, conn *rudp.Conn,
	bufs int, quit qu.C) (r *RUDP) {
	r = &RUDP{
		Keys:     idKeys,
		recvKeys: make(KeyChain),
		sendKeys: make(KeyChain),
		endpoint: remote,
		in:       make(ByteChan, bufs),
		conn:     conn,
		buffers:  make(RUDPBuffer),
		quit:     quit,
	}
	return
}

func (r *RUDP) Send(b slice.Bytes) (e error) {
	// todo: split into packets.
	if _, e = r.conn.Write(b); fails(e) {
	}
	return
}

func (r *RUDP) Receive() <-chan slice.Bytes { return r.in }
func (r *RUDP) Stop()                       { r.quit.Q() }
