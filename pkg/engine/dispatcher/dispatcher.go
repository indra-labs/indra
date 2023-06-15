package dispatcher

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/packet"
	"github.com/indra-labs/indra/pkg/engine/transport"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"go.uber.org/atomic"
)

const (
	// DefaultStartingParity is set to 64, or 25%
	DefaultStartingParity = 64
	// DefaultDispatcherRekey is 16mb to trigger rekey.
	DefaultDispatcherRekey = 1 << 20
	TimeoutPingCount       = 10
)

var (
	blue  = color.Blue.Sprint
	log   = log2.GetLogger()
	fails = log.E.Chk
)

type Completion struct {
	ID   nonce.ID
	Time time.Time
}

// Dispatcher is a message splitter/joiner and error correction adjustment system
// that aims to minimise message latency by trading it for bandwidth especially
// to cope with radio connections.
//
// In its initial implementation by necessity reliable network transports are
// used, which means that the message transit time is increased for packet
// retransmits, thus a longer transit time than the ping indicates packet
// transmit failures.
//
// PingDivergence is adjusted with each acknowledgement from the message transit
// time compared to the current ping, if it is within range of the ping RTT this
// doesn't affect the adjustment.
//
// DataSent / TotalSent provides the ratio of redundancy the channel is using.
// TotalSent is not from the parameters at send but from acknowledgements of
// how much data was received before a message was reconstructed. Thus, it is
// used in combination with the PingDivergence to recompute the Parity parameter
// used for adjusting error correction redundancy as each message is decoded.
type Dispatcher struct {
	// Parity is the parity parameter to use in packet segments, this value
	// should be adjusted up and down in proportion with collected data on how
	// many packets led to receipt against the ratio of DataSent / TotalSent *
	// PingDivergence.
	Parity atomic.Uint32
	// DataSent is the amount of actual data sent in messages.
	DataSent *big.Int
	// DataReceived is the amount of actual data received in messages.
	DataReceived *big.Int
	// TotalSent is the amount of bytes that were needed to successfully
	// transmit, including packet overhead. This is the raw size of the
	// segmented packets that were sent.
	TotalSent *big.Int
	// TotalReceived is the amount of bytes that were needed to successfully
	// transmit, including packet overhead. This is the raw size of the
	// segmented packets that were received.
	TotalReceived *big.Int
	ErrorEWMA     ewma.MovingAverage
	Ping          ewma.MovingAverage
	// PingDivergence represents the proportion of time between start of send
	// and receiving acknowledgement, versus the ping RTT being actively
	// measured concurrently. Shorter/equal time means it can reduce redundancy,
	// longer time needs to increase it.
	//
	// Combined with DataSent / TotalSent this guides the error correction
	// parameter for a given transmission that minimises latency. Onion routing
	// necessarily amplifies any latency so making a transmission get across
	// before/without retransmits is as good as the path can provide.
	PingDivergence  ewma.MovingAverage
	Duplex          *transport.DuplexByteChan
	Done            []Completion
	PendingInbound  []*RxRecord
	PendingOutbound []*TxRecord
	Partials        map[nonce.ID]packet.Packets
	Prv             []*crypto.Prv
	KeyLock         sync.Mutex
	lastRekey       *big.Int
	ks              *crypto.KeySet
	Conn            *transport.Conn
	Mutex           sync.Mutex
	Ready           qu.C
	ip              string
	rekeying        atomic.Bool
}

// GetRxRecordAndPartials returns the receive record and packets received so far
// for a message with a given ID>
func (d *Dispatcher) GetRxRecordAndPartials(id nonce.ID) (rxr *RxRecord,
	packets packet.Packets) {
	for _, v := range d.PendingInbound {
		if v.ID == id {
			rxr = v
			break
		}
	}
	var ok bool
	if packets, ok = d.Partials[id]; ok {
	}
	return
}

// Handle the message. This is expected to be called with the mutex locked,
// so nothing in it should be trying to lock it.
func (d *Dispatcher) Handle(m slice.Bytes, rxr *RxRecord) {
	for i := range d.Done {
		if d.Done[i].ID == rxr.ID {
			log.W.Ln(d.ip, "handle called for done message packet", rxr.ID)
			return
		}
	}
	hash := sha256.Single(m.ToBytes())
	copy(rxr.Hash[:], hash[:])
	s := splice.NewFrom(m)
	c := reg.Recognise(s)
	if c == nil {
		return
	}
	log.T.S(blue(d.Conn.LocalMultiaddr()) + " handling message") // m.ToBytes(),
	magic := c.Magic()
	log.T.Ln(d.ip, "decoding message with magic",
		color.Red.Sprint(magic))
	var e error
	if e = c.Decode(s); fails(e) {
		return
	}
	switch magic {
	case NewKeyMagic:
		o := c.(*NewKey)
		if d.Conn.GetRemoteKey().Equals(o.NewPubkey) {
			log.W.Ln(d.ip, "same key received again")
			return
		}
		d.Conn.SetRemoteKey(o.NewPubkey)
		log.D.Ln(d.ip, "new remote key received",
			o.NewPubkey.ToBased32())
	case AcknowledgeMagic:
		log.D.Ln("ack: received", len(d.Done))
		o := c.(*Acknowledge)
		r := o.RxRecord
		var tmp []*TxRecord
		for _, pending := range d.PendingOutbound {
			if pending.ID == r.ID {
				if r.Hash == pending.Hash {
					log.T.Ln("ack: accounting of successful send")
					d.DataSent = d.DataSent.Add(d.DataSent,
						big.NewInt(int64(pending.Size)))
				}
				log.T.Ln(blue(d.Conn.LocalMultiaddr()),
					d.ErrorEWMA.Value(), pending.Size, r.Size, r.Received,
					float64(pending.Size)/float64(r.Received))
				if pending.Size >= d.Conn.MTU-packet.Overhead {
					d.ErrorEWMA.Add(float64(pending.Size) / float64(r.Received))
				}
				log.T.Ln(d.ip, "first", pending.First.UnixNano(), "last",
					pending.Last.UnixNano(), r.Ping.Nanoseconds())
				tot := pending.Last.UnixNano() - pending.First.UnixNano()
				pn := r.Ping
				div := float64(pn) / float64(tot)
				log.T.Ln(d.ip, "div", div, "tot", tot)
				d.PingDivergence.Add(div)
				par := float64(d.Parity.Load())
				pv := par * d.PingDivergence.Value() *
					(1 + d.ErrorEWMA.Value())
				log.T.Ln(d.ip, "pv", par, "*", d.PingDivergence.Value(), "*",
					1+d.ErrorEWMA.Value(), "=", pv)
				d.Parity.Store(uint32(byte(pv)))
				log.T.Ln(d.ip, "ack: processed for",
					color.Green.Sprint(r.ID.String()),
					r.Ping, tot, div, par,
					d.PingDivergence.Value(),
					d.ErrorEWMA.Value(),
					d.Parity.Load())
				break
			} else {
				tmp = append(tmp, pending)
			}
		}
		// Entry is now deleted and processed.
		d.PendingOutbound = tmp
	case OnionMagic:
		o := c.(*Onion)
		d.Duplex.Receiver.Send(o.Bytes)
	}
}

// HandlePing adds a current ping result and combines it into the running
// exponential weighted moving average.
func (d *Dispatcher) HandlePing(p ping.Result) {
	d.Mx(func() (rtn bool) {
		d.Ping.Add(float64(p.RTT))
		return
	})
}

// Mx runs a closure with the dispatcher mutex locked which returns a bool that
// passes through to the result of the dispatcher.Mx function.
func (d *Dispatcher) Mx(fn func() bool) bool {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	return fn()
}

// ReKey sends a new key for the other party to use for sending future messages.
func (d *Dispatcher) ReKey() {
	d.lastRekey = d.lastRekey.SetBytes(d.TotalReceived.Bytes())
	if d.rekeying.Load() {
		log.D.Ln("trying to rekey while rekeying")
		return
	}
	d.rekeying.Toggle()
	defer func() {
		// time.Sleep(time.Second / 2)
		d.rekeying.Toggle()
	}()
	// log.I.Ln("rekey", d.TotalReceived, d.lastRekey)
	var e error
	var prv *crypto.Prv
	if prv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	rpl := NewKey{NewPubkey: crypto.DerivePub(prv)}
	keyMsg := splice.New(rpl.Len())
	if e = rpl.Encode(keyMsg); fails(e) {
		return
	}
	m := keyMsg.GetAll()
	id := nonce.NewID()
	hash := sha256.Single(m)
	txr := &TxRecord{
		ID:    id,
		Hash:  hash,
		First: time.Now(),
		Size:  len(m),
	}
	pp := &packet.PacketParams{
		ID:     id,
		To:     d.Conn.GetRemoteKey(),
		Parity: int(d.Parity.Load()),
		Length: m.Len(),
		Data:   m,
	}
	mtu := d.Conn.GetMTU()
	var packets [][]byte
	_, packets, e = packet.SplitToPackets(pp, mtu, d.ks)
	if fails(e) {
		return
	}
	cryptorand.Shuffle(len(packets), func(i, j int) {
		packets[i], packets[j] = packets[j], packets[i]
	})
	log.D.Ln(d.ip, "sending new key")
	sendChan := d.Conn.GetSend()
	for i := range packets {
		sendChan.Send(packets[i])
	}
	d.lastRekey = d.lastRekey.SetBytes(d.TotalReceived.Bytes())
	txr.Last = time.Now()
	txr.Ping = time.Duration(d.Ping.Value())
	for _, v := range packets {
		d.TotalSent = d.TotalSent.Add(d.TotalSent,
			big.NewInt(int64(len(v))))
	}
	d.PendingOutbound = append(d.PendingOutbound, txr)
	d.Prv = append(d.Prv, prv)
	if len(d.Prv) > 32 {
		d.Prv = d.Prv[:32]
	}
	log.D.Ln("previous keys", len(d.Prv))
}

// RecvFromConn receives a new message from the connection, checks if it can be
// reassembled and if it can, dispatches it to the receiver channel.
func (d *Dispatcher) RecvFromConn(m slice.Bytes) {
	log.T.Ln(color.Blue.Sprint(blue(d.Conn.LocalMultiaddr())), "received", len(m),
		"bytes from conn to dispatcher",
		// m.ToBytes(),
	)
	// Packet received, decrypt, gather and send acks back and reconstructed
	// messages to the Dispatcher.RecvFromConn channel.
	from, to, iv, e := packet.GetKeysFromPacket(m)
	if fails(e) {
		return
	}
	// This connection should only receive messages with cloaked keys
	// matching our private key of the connection.
	{
		log.T.Ln(d.ip, "keylock lock")
		d.KeyLock.Lock()
	}
	var prv *crypto.Prv
	for firstI := len(d.Prv) - 1; firstI >= 0; firstI-- {
		if !crypto.Match(to, crypto.DerivePub(d.Prv[firstI]).ToBytes()) {
			continue
		}
		prv = d.Prv[firstI]
	}
	func() {
		log.T.Ln(d.ip, "keylock unlock")
		d.KeyLock.Unlock()
	}()
	if prv == nil {
		log.D.Ln("did not find key for packet, discarding")
	}
	var p *packet.Packet
	log.T.Ln("decoding packet")
	if p, e = packet.DecodePacket(m, from, prv, iv); fails(e) {
		return
	}
	var rxr *RxRecord
	var packets packet.Packets
	d.Mx(func() (rtn bool) {
		d.TotalReceived = d.TotalReceived.Add(d.TotalReceived,
			big.NewInt(int64(len(m))))
		log.T.Ln("data since last rekey",
			big.NewInt(9).Sub(d.TotalReceived, d.lastRekey),
			d.lastRekey, d.TotalReceived, DefaultDispatcherRekey)
		if big.NewInt(0).Sub(d.TotalReceived,
			d.lastRekey).Uint64() > DefaultDispatcherRekey {
			d.lastRekey.SetBytes(d.TotalReceived.Bytes())
			d.ReKey()
		}
		return
	})
	if d.Mx(func() (rtn bool) {
		if rxr, packets = d.GetRxRecordAndPartials(p.ID); rxr != nil {
			log.T.Ln("more message", p.Length, len(p.Data))
			rxr.Received += uint64(len(m))
			rxr.Last = time.Now()
			log.T.Ln("rxr", rxr.Size)
			d.Partials[p.ID] = append(d.Partials[p.ID], p)
		} else {
			log.T.Ln("new message", p.Length, len(p.Data))
			rxr = &RxRecord{
				ID:       p.ID,
				First:    time.Now(),
				Last:     time.Now(),
				Size:     uint64(p.Length),
				Received: uint64(len(m)),
				Ping:     time.Duration(d.Ping.Value()),
			}
			d.PendingInbound = append(d.PendingInbound, rxr)
			packets = packet.Packets{p}
			d.Partials[p.ID] = packets
		}
		for i := range d.Done {
			if p.ID == d.Done[i].ID {
				log.T.Ln(blue(d.Conn.LocalMultiaddr()),
					"new packet from done message")
				// Skip, message has been dispatched.
				segCount := int(p.Length) / d.Conn.GetMTU()
				mod := int(p.Length) % d.Conn.GetMTU()
				if mod > 0 {
					segCount++
				}
				if len(d.Partials[p.ID]) >= segCount {
					log.T.Ln("deleting fully completed message")
					tmpD := make([]Completion, 0, len(d.PendingInbound))
					for i := range d.Done {
						if p.ID != d.Done[i].ID {
							tmpD = append(tmpD, d.Done[i])
						}
					}

					return true
				}
			}
		}
		return
	}) {
		return
	}
	log.T.Ln(blue(d.Conn.LocalMultiaddr()),
		"seq", p.Seq, int(p.Length), len(p.Data), p.ID, d.Done)
	segCount := int(p.Length) / d.Conn.GetMTU()
	mod := int(p.Length) % d.Conn.GetMTU()
	if mod > 0 {
		segCount++
	}
	var msg []byte
	if len(d.Partials[p.ID]) > segCount {
		if d.Mx(func() (rtn bool) {
			// Enough to attempt reconstruction:
			if d.Partials[p.ID], msg, e = packet.JoinPackets(d.Partials[p.ID]); fails(e) {
				log.D.Ln("failed to join packets")
				return
			}
			d.DataReceived = d.DataReceived.Add(d.DataReceived,
				big.NewInt(int64(len(msg))))
			for _, v := range d.Partials[p.ID] {
				if v == nil {
					continue
				}
				n := len(v.Data) + v.GetOverhead()
				d.TotalReceived = d.TotalReceived.Add(
					d.TotalReceived, big.NewInt(int64(n)),
				)
			}
			return
		}) {
			return
		}
		// Send the message on to the receiving channel.
		d.Handle(msg, rxr)
	}
}

// RunGC runs the garbage collection for the dispatcher. Stale data and completed
// transmissions are purged from memory.
func (d *Dispatcher) RunGC() {
	log.T.Ln(d.ip, "RunGC")
	// remove successful receives after all pieces arrived. Successful
	// transmissions before the timeout will already be cleared from
	// confirmation by the acknowledgment.
	var rxr []*RxRecord
	log.T.Ln(d.ip, "checking for stale partials")
	for dpi, dp := range d.Partials {
		// Find the oldest and newest.
		oldest := time.Now()
		if len(dp) == 0 || dp == nil {
			continue
		}
		var tmp packet.Packets
		for i := range dp {
			if dp[i] != nil {
				tmp = append(tmp, dp[i])
			}
		}
		dp = tmp
		newest := dp[0].TimeStamp
		for _, ts := range dp {
			if oldest.After(ts.TimeStamp) {
				oldest = ts.TimeStamp
			}
			if newest.Before(ts.TimeStamp) {
				newest = ts.TimeStamp
			}
		}
		if newest.Sub(time.Now()) > time.Duration(d.Ping.Value())*
			TimeoutPingCount {
			log.D.Ln("receive timed out with failure")
			// after 10 pings of time elapse since last received we
			// consider the transmission a failure, send back the
			// failed RxRecord and delete the map entry.
			var tmpR []*RxRecord
			for i := range d.PendingInbound {
				if d.PendingInbound[i].ID == dp[0].ID {
					rxr = append(rxr, d.PendingInbound[i])
				} else {
					tmpR = append(tmpR, d.PendingInbound[i])
				}
			}
			delete(d.Partials, dpi)
			tmpD := make([]Completion, 0, len(d.PendingInbound))
			d.PendingInbound = tmpR
			for i := range d.Done {
				if d.PendingInbound[i].ID == d.Done[i].ID {
					log.D.Ln("removing", dpi, d.Done[i].ID)
				} else {
					tmpD = append(tmpD, d.Done[i])
				}
			}
			d.Done = tmpD
		}
	}
	var e error
	for i := range rxr {
		// send the RxRecord to the peer.
		ack := &Acknowledge{rxr[i]}
		s := splice.New(ack.Len())
		if e = ack.Encode(s); fails(e) {
			continue
		}
		log.T.Ln(d.ip, "sending ack")
		d.Duplex.Send(s.GetAll())
	}
}

// SendAck sends an acknowledgement record for a successful transmission of a
// message.
func (d *Dispatcher) SendAck(rxr *RxRecord) {
	// Remove Rx from pending.
	log.T.Ln(d.ip, "mutex lock")
	d.Mutex.Lock()
	defer func() {
		log.T.Ln(d.ip, "mutex unlock")
		d.Mutex.Unlock()
	}()
	d.Mx(func() (r bool) {
		var tmp []*RxRecord
		for _, v := range d.PendingInbound {
			if rxr.ID != v.ID {
				tmp = append(tmp, v)
			} else {
				log.T.Ln(d.ip, "sending ack")
				ack := &Acknowledge{rxr}
				log.T.S(d.ip, "rxr size", rxr.Size)
				s := splice.New(ack.Len())
				_ = ack.Encode(s)
				d.Duplex.Send(s.GetAll())
			}
		}
		d.PendingInbound = tmp
		return
	})
}

// SendToConn delivers a buffer to be sent over the connection, and returns the
// number of packets that were sent.
func (d *Dispatcher) SendToConn(m slice.Bytes) (pieces int) {
	log.T.Ln(d.ip, "message dispatching to conn") // m.ToBytes(),

	// Data received for sending through the Conn.
	id := nonce.NewID()
	hash := sha256.Single(m)
	txr := &TxRecord{
		ID:    id,
		Hash:  hash,
		First: time.Now(),
		Size:  len(m),
	}
	pp := &packet.PacketParams{
		ID:     id,
		To:     d.Conn.GetRemoteKey(),
		Parity: int(d.Parity.Load()),
		Length: m.Len(),
		Data:   m,
	}
	mtu := d.Conn.GetMTU()
	var packets [][]byte
	var e error
	pieces, packets, e = packet.SplitToPackets(pp, mtu, d.ks)
	if fails(e) {
		return
	}
	// Shuffle. This is both for anonymity and improving the chances of most
	// error bursts to not cut through a whole segment (they are grouped by 256
	// for RS FEC).
	cryptorand.Shuffle(len(packets), func(i, j int) {
		packets[i], packets[j] = packets[j], packets[i]
	})
	// Send them out!
	sendChan := d.Conn.GetSend()
	for i := range packets {
		log.T.Ln(d.ip, "sending out", i)
		sendChan.Send(packets[i])
		log.T.Ln(d.ip, "sent", i)
	}
	txr.Last = time.Now()
	d.Mutex.Lock()
	txr.Ping = time.Duration(d.Ping.Value())
	for _, v := range packets {
		d.TotalSent = d.TotalSent.Add(d.TotalSent,
			big.NewInt(int64(len(v))))
	}
	d.PendingOutbound = append(d.PendingOutbound, txr)
	d.Mutex.Unlock()
	log.T.Ln(d.ip, "message dispatched")
	return
}

// RxRecord is the details of a message reception and mostly forms the data sent
// in a message received acknowledgement. This data goes into an acknowledgement
// message.
type RxRecord struct {
	ID nonce.ID
	// Hash is the hash of the reconstructed message received.
	Hash sha256.Hash
	// First is when the first packet was received.
	First time.Time
	// Last is when the last packet was received. A longer time than the current
	// ping RTT after First indicates retransmits.
	Last time.Time
	// Size of the message as found in the packet headers.
	Size uint64
	// Received is the number of bytes received upon reconstruction, including
	// packet overhead.
	Received uint64
	// Ping is the average ping RTT on the connection calculated at each packet
	// receive, used with the total message transmit time to estimate an
	// adjustment in the parity shards to be used in sending on this connection.
	Ping time.Duration
}

// TxRecord is the details of a send operation in progress. This is used with
// the data received in the acknowledgement, which is a completed RxRecord..
type TxRecord struct {
	ID nonce.ID
	// Hash is the record of the hash of the original message.
	sha256.Hash
	// First is the time the first piece was sent.
	First time.Time
	// Last is the time the last piece was sent.
	Last time.Time
	// Size is the number of bytes in the message payload.
	Size int
	// Ping is the recorded average current round trip time at send.
	Ping time.Duration
}

// NewDispatcher initialises and starts up a Dispatcher with the provided
// connection, acquired by dialing or Accepting inbound connection from a peer.
func NewDispatcher(l *transport.Conn, ctx context.Context,
	ks *crypto.KeySet) (d *Dispatcher) {
	d = &Dispatcher{
		Conn:           l,
		ks:             ks,
		Duplex:         transport.NewDuplexByteChan(transport.ConnBufs),
		Ping:           ewma.NewMovingAverage(),
		PingDivergence: ewma.NewMovingAverage(),
		ErrorEWMA:      ewma.NewMovingAverage(),
		DataReceived:   big.NewInt(0),
		DataSent:       big.NewInt(0),
		TotalSent:      big.NewInt(0),
		TotalReceived:  big.NewInt(0),
		lastRekey:      big.NewInt(0),
		Partials:       make(map[nonce.ID]packet.Packets),
		Ready:          qu.T(),
	}
	d.rekeying.Store(false)
	d.ip = blue(d.Conn.RemoteMultiaddr())
	var e error
	prk := d.Conn.RemotePublicKey()
	var rprk slice.Bytes
	if rprk, e = prk.Raw(); fails(e) {
		return
	}
	cpr := crypto.PrvKeyFromBytes(rprk)
	d.Prv = append(d.Prv, cpr)
	fpk := d.Conn.RemotePublicKey()
	var rpk slice.Bytes
	if rpk, e = fpk.Raw(); fails(e) {
		return
	}
	var pk *crypto.Pub
	if pk, e = crypto.PubFromBytes(rpk); fails(e) {
		return
	}
	d.Conn.SetRemoteKey(pk)
	d.PingDivergence.Set(1)
	d.ErrorEWMA.Set(0)
	d.Parity.Store(DefaultStartingParity)
	ps := ping.NewPingService(l.Host)
	pings := ps.Ping(ctx, l.Conn.RemotePeer())
	garbageTicker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-d.Ready.Wait():
				for {
					select {
					case <-garbageTicker.C:
						d.RunGC()
					case p := <-pings:
						d.HandlePing(p)
					case m := <-d.Conn.Transport.Receive():
						d.RecvFromConn(m)
					case m := <-d.Duplex.Sender.Receive():
						d.SendToConn(m)
					case <-ctx.Done():
						return
					}
				}
			case <-garbageTicker.C:
			case p := <-pings:
				d.HandlePing(p)
			case m := <-d.Conn.Transport.Receive():
				d.RecvFromConn(m)
			case <-ctx.Done():
				return
			}
		}
	}()
	// Start key exchange.
	d.Mx(func() bool {
		d.ReKey()
		return false
	})
	d.Ready.Q()
	return
}
