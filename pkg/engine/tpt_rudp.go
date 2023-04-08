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
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/rudp"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	RCPHandshakeLen = nonce.IVLen + 2*pub.KeyLen
	KeychangeMagic  = "KEYC"
)

type (
	RCPBuffer map[nonce.ID][]*Packet
	KeySlot   struct {
		*Keys
		lastUsed time.Time
	}
	KeyChain map[string]*KeySlot
	// RCP - Resilient Communication Protocol is a protocol based on rUDP from
	// Plan 9 OS that adds dynamic error correction redundancy to messages
	// aiming to avoid retransmission, minimising latency by trading off
	// bandwidth capacity.
	RCP struct {
		*Keys
		mtu                int
		recvKeys, sendKeys KeyChain
		endpoint           *netip.AddrPort
		in                 ByteChan
		uConn              *net.UDPConn
		conn               *rudp.Conn
		good, corrupt      atomic.Uint64
		failRate           atomic.Uint32
		bufMx              sync.Mutex
		buffers            RCPBuffer
		ks                 *signer.KeySet
		quit               qu.C
	}
)

func NewListenerRCP(idKeys *Keys, address string, bufs,
	mtu int, ks *signer.KeySet, quit qu.C) (r *RCP, e error) {
	
	if mtu <= 512 {
		// Conservative packet size limit.
		mtu = 1382
	}
	bindAddress := net.ParseIP(address)
	network := "udp"
	if bindAddress.To4() != nil {
		network = "udp4"
	}
	addr := &net.UDPAddr{IP: bindAddress, Port: 0}
	var conn *net.UDPConn
	if conn, e = net.ListenUDP(network, addr); fails(e) {
		return
	}
	r = MakeRCP(idKeys, nil, conn, bufs, mtu, ks, quit)
	buf := slice.NewBytes(mtu)
	go r.listen(buf)
	return
}

func NewOutboundRCP(idKeys *Keys, remote string,
	rKey *pub.Key, local string, bufs, mtu int, ks *signer.KeySet,
	quit qu.C) (r *RCP, e error) {
	
	log.D.Ln("opening connection to", remote, "from", local)
	if mtu <= 512 {
		mtu = 1382 // Conservative packet size limit.
	}
	bindAddress := net.ParseIP(local)
	network := "udp"
	if bindAddress.To4() == nil {
		network = "udp4"
	}
	var ap netip.AddrPort
	if ap, e = netip.ParseAddrPort(remote); fails(e) {
		return
	}
	raddr, laddr := net.UDPAddrFromAddrPort(ap),
		&net.UDPAddr{IP: bindAddress}
	var conn *net.UDPConn
	if conn, e = net.DialUDP(network, laddr, raddr); fails(e) {
		return
	}
	
	r = MakeRCP(idKeys, &ap, conn, bufs, mtu, ks, quit)
	log.D.Ln("connection open", r.conn.LocalAddr(), "->", r.conn.RemoteAddr())
	var recvKey, ciphKey *Keys
	if recvKey, ciphKey, e = Generate2Keys(); fails(e) {
		return
	}
	reply, iv := NewSplice(RCPHandshakeLen), nonce.New()
	reply.Pubkey(ciphKey.Pub).IV(iv)
	start := reply.GetCursor()
	reply.Pubkey(recvKey.Pub)
	if fails(Encipher(reply.GetFrom(start), iv, ciphKey.Prv, rKey,
		"handshake encode")) {
		return
	}
	var n int
	// Send the connection receiver public key to the other side.
	log.D.S("writing data to "+remote, reply.GetAll().ToBytes())
	if n, e = r.conn.Write(reply.GetAll()); fails(e) ||
		n != RCPHandshakeLen {
		return
	}
	addr := raddr.String()
	r.Mx(func() {
		r.recvKeys[addr] = LoadKeySlot(recvKey.Prv, recvKey.Pub)
		// Populate so the listener loads the reply key in here:
		r.sendKeys[addr] = LoadKeySlot(nil, nil)
	})
	buf := slice.NewBytes(mtu)
	go r.listen(buf)
	return
}

func (r *RCP) Send(b slice.Bytes) (e error) {
	pp := &PacketParams{
		ID:     nonce.NewID(),
		To:     r.recvKeys[r.endpoint.String()].Pub,
		From:   r.ks.Next(),
		Parity: 24,
		Seq:    0,
		Length: len(b),
		Data:   b,
	}
	var bytes [][]byte
	_, bytes, e = SplitToPackets(pp, r.mtu)
	for i := range bytes {
		log.D.S("sending", r.endpoint.String(), bytes[i])
		if _, e = r.conn.Write(bytes[i]); fails(e) {
			return
		}
	}
	log.D.S("sent")
	return
}
func (r *RCP) Receive() <-chan slice.Bytes { return r.in }
func (r *RCP) Stop()                       { r.quit.Q() }

func MakeRCP(idKeys *Keys, remote *netip.AddrPort, conn *net.UDPConn,
	bufs, mtu int, ks *signer.KeySet, quit qu.C) (r *RCP) {
	r = &RCP{
		Keys:     idKeys,
		mtu:      mtu,
		recvKeys: make(KeyChain),
		sendKeys: make(KeyChain),
		endpoint: remote,
		in:       make(ByteChan, bufs),
		uConn:    conn,
		conn:     rudp.NewConn(conn, rudp.New()),
		buffers:  make(RCPBuffer),
		ks:       ks,
		quit:     quit,
	}
	return
}

func (r *RCP) listen(buf slice.Bytes) {
	
	var e error
	listener := rudp.NewListener(r.uConn)
out:
	for {
		log.T.F("starting RCP listener for %s %s", r.conn.LocalAddr(),
			r.Keys.Pub.ToBase32())
		total := r.good.Load() + r.corrupt.Load()
		// Compute average out of 256 as 100% vs 0% and average
		// with previous failRate.
		ratio := 256 * total / (1 + r.corrupt.Load()) // avoiding divide by zero
		r.failRate.Store(uint32(byte((ratio +
			uint64(r.failRate.Load())) / 2)))
		if total > 1<<24 {
			// After 16mb of traffic we reset the counters.
			r.good.Store(0)
			r.corrupt.Store(0)
		}
		var rConn *rudp.Conn
		log.D.Ln("accepting connections at", listener.Addr().String())
		if rConn, e = listener.AcceptRudp(); fails(e) {
			break
		}
		log.D.Ln("accepting connection on", rConn.LocalAddr().String())
		var n int
		addr := rConn.RemoteAddr().String()
		if n, e = rConn.Read(buf); fails(e) {
			if e.Error() == "corrupt" {
				r.corrupt.Add(uint64(n))
				continue
			}
			continue
		}
		log.D.Ln("read", n, "bytes")
		r.good.Add(uint64(n))
		s := NewSpliceFrom(buf[:n])
		log.D.S("splice", s.GetAll().ToBytes())
		var magic string
		s.ReadMagic4(&magic)
		switch magic {
		case PacketMagic:
			log.D.S("incoming", s.GetAll())
			var fromPub *pub.Key
			var toCloak cloak.PubKey
			var iv nonce.IV
			if fromPub, toCloak, iv, e = GetKeysFromPacket(s.GetAll()); fails(e) {
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
			// The minimum pieces need to recover the packet are computable by
			// the length field and the length of the data field. If the number
			// in the buffer exceeds the data shards required we want to attempt
			// reassembly.
			lq := int(pkt.Length) / len(pkt.Data)
			lm := int(pkt.Length) % len(pkt.Data)
			if lm != 0 {
				lq++
			}
			var buffers []*Packet
			var msg []byte
			r.Mx(func() {
				r.buffers[pkt.ID] = append(r.buffers[pkt.ID], pkt)
				if len(r.buffers[pkt.ID]) >= lq {
					buffers = r.buffers[pkt.ID]
				}
				if buffers != nil {
					if buffers, msg, e = JoinPackets(buffers); fails(e) {
						// Pass it back afterwards as the JoinPackets function
						// cleans stuff that doesn't need to be done twice.
						r.buffers[pkt.ID] = buffers
					}
				}
			})
			if msg != nil {
				r.in <- msg
			}
			// todo: handling stale stuff that never decoded.
		case KeychangeMagic:
			log.D.S("incoming", s.GetAll())
			if s.Len() < 4+pub.KeyLen {
				log.D.F("message too short to contain a public key")
				continue
			}
			var peerKey *pub.Key
			s.ReadPubkey(&peerKey)
			if peerKey == nil {
				continue
			}
			r.Mx(func() {
				if _, ok := r.sendKeys[addr]; ok {
					r.sendKeys[addr] = LoadKeySlot(nil, peerKey)
				}
			})
		default:
			log.D.S("incoming", s.GetAll())
			// If it is not a message, it is the delivery of a receiver public
			// key for the client, and we generate a new one and send it back
			// encrypted in a packet addressed to the provided key.
			if n < RCPHandshakeLen {
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
			o := NewSplice(pub.KeyLen + 4).
				Magic4(KeychangeMagic).
				Pubkey(recvPub)
			var pkt slice.Bytes
			if pkt, e = EncodePacket(&PacketParams{
				To:     pK,
				From:   recvPrv,
				Parity: int(r.GetParity()),
				Length: o.Len(),
				Data:   o.GetAll(),
			}); fails(e) {
				continue
			}
			if n, e = r.uConn.Write(pkt); fails(e) &&
				n != len(pkt) {
				r.corrupt.Add(uint64(n))
				continue
			}
			r.good.Add(uint64(n))
			log.D.Ln("sent key change message reply")
		}
		select {
		case <-r.quit:
			if e = rConn.Close(); fails(e) {
			}
			break out
		default:
		}
	}
	log.W.F("stopped RCP listener for %s", r.uConn.LocalAddr().String())
}

func (r *RCP) Mx(fn func()) {
	r.bufMx.Lock()
	fn()
	r.bufMx.Unlock()
}

func (r *RCP) GetParity() (parity byte) {
	fr := r.failRate.Load()
	p := fr * 10 / 11 // 10% higher than fail rate.
	if p > 255 {
		p = 255
	}
	parity = byte(256 - p)
	return
}

func LoadKeySlot(pr *prv.Key, pb *pub.Key) (k *KeySlot) {
	var pBytes pub.Bytes
	if pb != nil {
		pBytes = pb.ToBytes()
	}
	return &KeySlot{&Keys{Pub: pb, Bytes: pBytes, Prv: pr}, time.Now()}
}
