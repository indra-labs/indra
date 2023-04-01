package engine

import (
	"crypto/cipher"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/packet"
	"git-indra.lan/indra-labs/indra/pkg/rudp"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	RUDPHandshakeLen = nonce.IVLen + 2*pub.KeyLen
	KeychangeMagic   = "KEYC"
)

type RUDPBuffer map[nonce.IV][]slice.Bytes

type KeySlot struct {
	*Keys
	lastUsed time.Time
}

type KeyChain map[string]*KeySlot

type RUDP struct {
	identityKeys       *Keys
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

func NewOutboundRUDP(idKeys *Keys, remote *netip.AddrPort,
	remoteKey *pub.Key, local net.IP, bufs, mtu int, quit qu.C) (r *RUDP,
	e error) {
	
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
		identityKeys: idKeys,
		recvKeys:     make(KeyChain),
		sendKeys:     make(KeyChain),
		endpoint:     remote,
		in:           make(ByteChan, bufs),
		conn:         rUDPConn,
		buffers:      make(RUDPBuffer),
		quit:         quit,
	}
	// Key exchange handshake.
	var sendPrv, cipherPrv *prv.Key
	if sendPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	sendPub := pub.Derive(sendPrv)
	if cipherPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	cipherPub := pub.Derive(sendPrv)
	sendPubBytes := sendPub.ToBytes()
	handshakeReply := NewSplice(RUDPHandshakeLen)
	iv := nonce.New()
	handshakeReply.IV(iv).Pubkey(cipherPub)
	start := handshakeReply.GetCursor()
	handshakeReply.Pubkey(sendPub)
	var blk cipher.Block
	if blk = ciph.GetBlock(cipherPrv, remoteKey,
		"handshake encode"); check(e) {
		return
	}
	ciph.Encipher(blk, iv, handshakeReply.GetFrom(start))
	var n int
	// Send the connection receiver public key to the other side.
	if n, e = r.conn.Write(handshakeReply.GetAll()); check(e) &&
		n != len(sendPubBytes) {
		return
	}
	buf := make(slice.Bytes, mtu)
	// Other side should respond with a Packet containing their receiver
	// public key.
	if _, e = r.conn.Read(buf); check(e) {
		return
	}
	
	go r.listen(conn, buf, quit)
	return
}

func NewListenerRUDP(idKeys *Keys, bindAddress net.IP, bufs,
	mtu int, quit qu.C) (r *RUDP, e error) {
	
	if mtu <= 0 {
		// Conservative packet size limit.
		mtu = 1382
	}
	addr := &net.UDPAddr{IP: bindAddress, Port: 0}
	var conn *net.UDPConn
	conn, e = net.ListenUDP("udp", addr)
	if check(e) {
		return
	}
	rUDPConn := rudp.NewConn(conn, rudp.New())
	r = &RUDP{
		identityKeys: idKeys,
		recvKeys:     make(KeyChain),
		in:           make(ByteChan, bufs),
		conn:         rUDPConn,
		buffers:      make(RUDPBuffer),
		quit:         quit,
	}
	buf := make(slice.Bytes, mtu)
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
		x := 256*total/(1+r.corrupt.Load()) - 1 // avoiding divide by zero
		r.failRate.Store(uint32(byte((x + uint64(r.failRate.Load())) / 2)))
		if total > 1<<24 {
			// After 16mb of traffic we reset the counters.
			r.good.Store(0)
			r.corrupt.Store(0)
		}
		var rConn *rudp.Conn
		if rConn, e = listener.AcceptRudp(); check(e) {
			break
		}
		var n int
		addr := rConn.RemoteAddr().String()
		if n, e = rConn.Read(buf); check(e) {
			if e.Error() == "corrupt" {
				r.corrupt.Add(uint64(n))
				continue
			}
			continue
		}
		r.good.Add(uint64(n))
		s := LoadSplice(buf[:n], slice.NewCursor())
		var magic string
		s.ReadMagic4(&magic)
		switch magic {
		case packet.Magic:
			var fromPub *pub.Key
			var toCloak cloak.PubKey
			if fromPub, toCloak, e = packet.GetKeys(s.GetAll()); check(e) {
				continue
			}
			_ = fromPub
			_ = toCloak
			// todo: collect and manage packet buffers.
			r.in <- s.GetAll()
		case KeychangeMagic:
		default:
			if n < RUDPHandshakeLen {
				log.D.Ln("message too short for handshake")
				return
			}
			s.SetCursor(s.GetCursor() - 4)
			// If it is not a message, it is the delivery of a receiver public
			// key for the client, and we generate a new one and send it back
			// encrypted in a packet.
			var iv nonce.IV
			// The message format is: initialisation vector, sender public key,
			// receiver key is the identity key of the node.
			s.ReadIV(&iv)
			var pubKey *pub.Key
			s.ReadPubkey(&pubKey)
			if pubKey == nil {
				continue
			}
			var blk cipher.Block
			if blk = ciph.GetBlock(r.identityKeys.Prv, pubKey,
				"handshake decode"); check(e) {
				continue
			}
			ciph.Encipher(blk, iv, s.GetCursorToEnd())
			var peerKey *pub.Key
			s.ReadPubkey(&peerKey)
			r.bufMx.Lock()
			r.sendKeys[addr] = &KeySlot{
				Keys: &Keys{
					Pub:   peerKey,
					Bytes: peerKey.ToBytes(),
				},
				lastUsed: time.Now(),
			}
			r.bufMx.Unlock()
			var recvPrv *prv.Key
			if recvPrv, e = prv.GenerateKey(); check(e) {
				return
			}
			recvPub := pub.Derive(recvPrv)
			r.bufMx.Lock()
			r.recvKeys[addr] = &KeySlot{
				Keys: &Keys{
					Pub:   recvPub,
					Bytes: recvPub.ToBytes(),
					Prv:   recvPrv,
				},
				lastUsed: time.Now(),
			}
			r.bufMx.Unlock()
			
		}
		select {
		case <-quit:
			if e = rConn.Close(); check(e) {
			}
			return
		default:
		}
	}
}

func (r *RUDP) Send(b slice.Bytes) (e error) {
	// todo: split into packets.
	if _, e = r.conn.Write(b); check(e) {
	}
	return
}

func (r *RUDP) Receive() <-chan slice.Bytes { return r.in }
func (r *RUDP) Stop()                       { r.quit.Q() }
