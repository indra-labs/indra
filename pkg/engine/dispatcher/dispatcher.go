package dispatcher

import (
	"context"
	"math/big"
	"sync"
	"time"
	
	"github.com/VividCortex/ewma"
	"github.com/cybriq/qu"
	"github.com/gookit/color"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/engine/packet"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

var blue = color.Blue.Sprint

const (
	// DefaultStartingParity is set to 64, or 25%
	DefaultStartingParity = 64
	// DefaultDispatcherRekey is 16mb to trigger rekey.
	DefaultDispatcherRekey = 1 << 20
)

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
	PingDivergence        ewma.MovingAverage
	Duplex                *transport.DuplexByteChan
	Done                  []Completion
	PendingInbound        []*RxRecord
	PendingOutbound       []*TxRecord
	Partials              map[nonce.ID]packet.Packets
	Prv, OldPrv, OlderPrv *crypto.Prv
	KeyLock               sync.Mutex
	lastRekey             atomic.Value
	ks                    *crypto.KeySet
	Conn                  *transport.Conn
	Mutex                 sync.Mutex
	Ready                 qu.C
}

const TimeoutPingCount = 10

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
		Partials:       make(map[nonce.ID]packet.Packets),
		Ready:          qu.Ts(1),
	}
	var e error
	prk := d.Conn.Conn.LocalPrivateKey()
	var rprk slice.Bytes
	if rprk, e = prk.Raw(); fails(e) {
		return
	}
	cpr := crypto.PrvKeyFromBytes(rprk)
	d.Prv = cpr
	d.OldPrv = cpr
	d.OlderPrv = cpr
	d.lastRekey.Store(d.TotalReceived.Bytes())
	fpk := d.Conn.Conn.RemotePublicKey()
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
			case <-garbageTicker.C:
				d.RunGC()
			case p := <-pings:
				d.HandlePing(p)
			case m := <-l.Transport.Receive():
				d.RecvFromConn(m)
			case <-d.Ready.Wait():
				select {
				case m := <-d.Duplex.Sender.Receive():
					d.SendToConn(m)
				default:
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	// Start key exchange.
	d.KeyExchange()
	
	return
}

func (d *Dispatcher) RunGC() {
	log.T.Ln(d.Conn.RemoteMultiaddr(), "RunGC")
	// remove successful receives after all pieces arrived. Successful
	// transmissions before the timeout will already be cleared from
	// confirmation by the acknowledgment.
	var rxr []*RxRecord
	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	d.Mutex.Lock()
	d.Ready.Signal()
	// log.I.S("dispatcher state", d)
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
	// If we have moved a lot of data, time to rekey.
	last := big.NewInt(0).SetBytes(d.lastRekey.Load().([]byte))
	if last.Sub(last, d.TotalReceived).Uint64() > DefaultDispatcherRekey {
		d.KeyExchange()
	}
	func() {
		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
		d.Mutex.Unlock()
	}()
	// With the lock released we now can dispatch the RxRecords.
	var e error
	for i := range rxr {
		// send the RxRecord to the peer.
		ack := &Acknowledge{rxr[i]}
		s := splice.New(ack.Len())
		if e = ack.Encode(s); fails(e) {
			continue
		}
		d.Duplex.Sender.Send(s.GetAll())
	}
}

func (d *Dispatcher) HandlePing(p ping.Result) {
	// log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	d.Mutex.Lock()
	d.Ping.Add(float64(p.RTT))
	// func() {
	// log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
	d.Mutex.Unlock()
	// }()
}

func (d *Dispatcher) RecvFromConn(m slice.Bytes) {
	// d.Mutex.Lock()
	// defer func() {
	// 	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
	// 	d.Mutex.Unlock()
	// 	// <-d.Ready // .Signal()
	// }()
	log.T.Ln(color.Blue.Sprint(d.Conn.LocalMultiaddr()), "received", len(m),
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
	// d.Mutex.Lock()
	// log.D.S(to)
	{
		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock lock")
		d.KeyLock.Lock()
	}
	prv := d.Prv
	if !crypto.Match(to, crypto.DerivePub(prv).ToBytes()) {
		log.W.Ln(d.Conn.LocalMultiaddr(), "first", crypto.DerivePub(prv))
		prv = d.OldPrv
		if !crypto.Match(to, crypto.DerivePub(prv).ToBytes()) {
			log.W.Ln(d.Conn.LocalMultiaddr(), "second")
			prv = d.OlderPrv
			if !crypto.Match(to, crypto.DerivePub(prv).ToBytes()) {
				func() {
					log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock unlock")
					d.KeyLock.Unlock()
				}()
				log.W.Ln(d.Conn.LocalMultiaddr(), "third")
				// debug.PrintStack()
				panic("why not!!!!")
				return
			}
		}
	}
	func() {
		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock unlock")
		d.KeyLock.Unlock()
	}()
	var p *packet.Packet
	log.T.Ln("decoding packet")
	if p, e = packet.DecodePacket(m, from, prv, iv); fails(e) {
		return
	}
	// log.T.Ln("decoded packet", p.ID)
	var rxr *RxRecord
	var packets packet.Packets
	// log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	// d.Mutex.Lock()
	d.TotalReceived = d.TotalReceived.Add(d.TotalReceived,
		big.NewInt(int64(len(m))))
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
			log.T.Ln(d.Conn.Conn.LocalMultiaddr(),
				"new packet from done message")
			// Skip, message has been dispatched.
			segCount := int(p.Length) / d.Conn.GetMTU()
			mod := int(p.Length) % d.Conn.GetMTU()
			if mod > 0 {
				segCount++
			}
			if len(d.Partials[p.ID]) >= segCount {
				log.T.Ln("deleting fully completed message")
				// wasDone := false
				tmpD := make([]Completion, 0, len(d.PendingInbound))
				for i := range d.Done {
					if p.ID != d.Done[i].ID {
						tmpD = append(tmpD, d.Done[i])
						// } else {
						// 	wasDone = true
					}
				}
				// {
				// 	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
				// d.Mutex.Unlock()
				// }
				return
			}
		}
	}
	// {
	// 	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
	// 	d.Mutex.Unlock()
	// }
	
	// // if the message is only one packet we can hand it on to the receiving
	// // channel now:
	// if (p.Seq == 0 || p.Seq == 1) &&
	// 	len(d.Partials[p.ID]) < 1 &&
	// 	int(p.Length) <= len(p.Data) {
	//
	// 	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	// 	d.Mutex.Lock()
	// log.T.Ln("forwarding single packet message")
	// d.Handle(m, rxr)
	// d.Done = append(d.Done, Completion{rxr.ID, time.Now()})
	// 	{
	// 		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
	// 		d.Mutex.Unlock()
	// 	}
	// 	return
	// }
	// for i := range d.Done {
	// 	if d.Done[i].ID == p.ID {
	// 		log.D.Ln("is done", p.ID)
	// 		return
	// 	}
	// }
	log.T.Ln(blue(d.Conn.Conn.LocalMultiaddr()),
		"seq", p.Seq, int(p.Length), len(p.Data), p.ID, d.Done)
	segCount := int(p.Length) / d.Conn.GetMTU()
	mod := int(p.Length) % d.Conn.GetMTU()
	if mod > 0 {
		segCount++
	}
	// log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	// d.Mutex.Lock()
	if len(d.Partials[p.ID]) > segCount {
		// Enough to attempt reconstruction:
		var msg []byte
		if d.Partials[p.ID], msg, e = packet.JoinPackets(d.Partials[p.ID]); fails(e) {
			log.D.Ln("failed to join packets")
			return
		}
		// log.T.S("message", msg)
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
		// Send the message on to the receiving channel.
		// d.Done = append(d.Done,
		// 	Completion{rxr.ID, time.Now()})
		d.Handle(msg, rxr)
		log.T.Ln(d.Conn.LocalMultiaddr(), "partials", len(d.Partials[p.ID]),
			"parity", d.Parity.Load())
		pars := segCount * int(d.Parity.Load()) / 256
		if d.Parity.Load() > 0 && pars < 1 {
			pars = 1
		}
		// // If we have all the pieces and handled it, it can be deleted.
		// if len(d.Partials[p.ID]) >= segCount+pars {
		// 	log.D.Ln("deleting completed message")
		// 	wasDone := false
		// 	tmpD := make([]Completion, 0, len(d.PendingInbound))
		// 	for i := range d.Done {
		// 		if p.ID != d.Done[i].ID {
		// 			tmpD = append(tmpD, d.Done[i])
		// 		} else {
		// 			wasDone = true
		// 		}
		// 	}
		// 	d.Done = tmpD
		// 	if !wasDone {
		// 		//
		// 		return
		// 	}
		// 	log.D.Ln("parity", segCount, pars, segCount+pars,
		// 		len(d.Partials[p.ID]))
		// 	// Remove the pending record for the complete tx.
		// 	var tmp []*RxRecord
		// 	for i := range d.PendingInbound {
		// 		if d.PendingInbound[i].ID != p.ID {
		// 			tmp = append(tmp, d.PendingInbound[i])
		// 		}
		// 	}
		// 	d.PendingInbound = tmp
		// 	delete(d.Partials, p.ID)
		// 	return
		// }
		// } else {
		// 	// Message tx is complete, but there should be more pieces to come or a
		// 	// timeout.
		// 	log.D.Ln("when is the future")
		// 	d.Done = append(d.Done,
		// 		Completion{rxr.ID, time.Now()})
	}
	// d.Mutex.Unlock()
}

func (d *Dispatcher) SendAck(rxr *RxRecord) {
	// log.T.S("sent ack", ack)
	// Remove Rx from pending.
	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	d.Mutex.Lock()
	defer func() {
		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
		d.Mutex.Unlock()
	}()
	var tmp []*RxRecord
	for _, v := range d.PendingInbound {
		if rxr.ID != v.ID {
			tmp = append(tmp, v)
			// } else {
			// 	log.T.Ln("removed pending receive record")
		} else {
			{
				log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex unlock")
				d.Mutex.Unlock()
			}
			log.T.Ln("sending ack")
			ack := &Acknowledge{rxr}
			log.T.S("rxr size", rxr.Size)
			s := splice.New(ack.Len())
			_ = ack.Encode(s)
			// d.Ready.Signal()
			// <-d.Ready
			d.Duplex.Send(s.GetAll())
			log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
			d.Mutex.Lock()
		}
	}
	d.PendingInbound = tmp
	// <-d.Ready
}

func (d *Dispatcher) SendToConn(m slice.Bytes) {
	log.T.S(d.Conn.LocalMultiaddr().String()+" message dispatching to conn",
		m.ToBytes())
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
	// log.D.Ln("completed", len(d.Done))
	var packets [][]byte
	var e error
	packets, e = packet.SplitToPackets(pp, mtu, d.ks)
	if fails(e) {
		return
	}
	// log.D.S("split packets", packets)
	// Shuffle. This is both for anonymity and improving the chances of most
	// error bursts to not cut through a whole segment (they are grouped by 256
	// for RS FEC).
	cryptorand.Shuffle(len(packets), func(i, j int) {
		packets[i], packets[j] = packets[j], packets[i]
	})
	// Send them out!
	sendChan := d.Conn.GetSend()
	for i := range packets {
		log.T.Ln("sending out", i)
		sendChan.Send(packets[i])
		log.T.Ln("sent", i)
		// d.Ready.Signal()
	}
	txr.Last = time.Now()
	// log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "mutex lock")
	d.Mutex.Lock()
	txr.Ping = time.Duration(d.Ping.Value())
	d.Mutex.Unlock()
	for _, v := range packets {
		d.TotalSent = d.TotalSent.Add(d.TotalSent,
			big.NewInt(int64(len(v))))
	}
	d.PendingOutbound = append(d.PendingOutbound, txr)
	log.T.Ln("message dispatched")
	// d.Ready.Signal()
}

func (d *Dispatcher) KeyExchange() {
	log.D.Ln(d.Conn.LocalMultiaddr(), "initiating key exchange")
	var e error
	var prv *crypto.Prv
	if prv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	rpl := NewKey{NewPubkey: crypto.DerivePub(prv)}
	reply := splice.New(rpl.Len())
	if e = rpl.Encode(reply); fails(e) {
		return
	}
	// d.Ready.Signal()
	d.Duplex.Send(reply.GetAll())
	log.D.Ln(d.Conn.LocalMultiaddr(), "initiating key exchange")
	// time.Sleep(time.Second)
	log.D.Ln("sending kx")
	{
		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock lock")
		d.KeyLock.Lock()
	}
	log.D.Ln("updating prv keys")
	d.OlderPrv = d.OldPrv
	d.OldPrv = d.Prv
	d.Prv = prv
	func() {
		log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock unlock")
		d.KeyLock.Unlock()
	}()
	d.lastRekey.Store(d.TotalReceived.Bytes())
	// d.Ready.Signal()
	// d.Ready.Q()
}

// Handle the message. This is expected to be called with the mutex locked,
// so nothing in it should be trying to lock it.
func (d *Dispatcher) Handle(m slice.Bytes, rxr *RxRecord) {
	for i := range d.Done {
		if d.Done[i].ID == rxr.ID {
			log.W.Ln("handle called for done message packet", rxr.ID)
			return
		}
	}
	// d.Done = append(d.Done, Completion{rxr.ID, time.Now()})
	// unlock := func() {
	// 	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock unlock")
	// 	d.KeyLock.Unlock()
	// }
	hash := sha256.Single(m.ToBytes())
	copy(rxr.Hash[:], hash[:])
	s := splice.NewFrom(m)
	c := onions.Recognise(s)
	if c == nil {
		return
	}
	log.T.S(d.Conn.LocalMultiaddr().String()+" handling message",
		m.ToBytes(),
	)
	magic := c.Magic()
	log.D.Ln("decoding message with magic", color.Red.Sprint(magic))
	var e error
	if e = c.Decode(s); fails(e) {
		return
	}
	switch magic {
	case NewKeyMagic:
		// {
		// 	log.T.Ln(d.Conn.Conn.LocalMultiaddr(), "keylock lock")
		// 	d.KeyLock.Lock()
		// }
		// defer unlock()
		o := c.(*NewKey)
		// log.D.Ln(d.Conn.LocalMultiaddr(), "new key")
		// var prv *crypto.Prv
		// if prv, e = crypto.GeneratePrvKey(); fails(e) {
		// 	unlock()
		// 	return
		// }
		// d.OlderPrv = d.OldPrv
		// d.OldPrv = d.Prv
		// log.D.Ln("changing connection key")
		// d.Prv = prv
		// d.lastRekey.Store(d.TotalReceived.Bytes())
		if d.Conn.GetRemoteKey().Equals(o.NewPubkey) {
			log.W.Ln("same key received again")
			return
		}
		d.Conn.SetRemoteKey(o.NewPubkey)
		log.D.Ln(blue(d.Conn.Conn.LocalMultiaddr()), "new remote key received",
			o.NewPubkey.ToBase32())
		// d.Done = append(d.Done, Completion{rxr.ID, time.Now()})
		// d.Ready.Signal()
		log.D.S("done", d.Done)
	case AcknowledgeMagic:
		log.D.Ln("ack: received", len(d.Done))
		o := c.(*Acknowledge)
		r := o.RxRecord
		// log.D.S("RxRecord", r, r.Size)
		var tmp []*TxRecord
		for _, pending := range d.PendingOutbound {
			if pending.ID == r.ID {
				if r.Hash == pending.Hash {
					log.T.Ln("ack: accounting of successful send")
					d.DataSent = d.DataSent.Add(d.DataSent,
						big.NewInt(int64(pending.Size)))
				}
				log.T.Ln(d.ErrorEWMA.Value(), pending.Size, r.Size, r.Received,
					float64(pending.Size)/float64(r.Received))
				if pending.Size >= d.Conn.MTU-packet.Overhead {
					d.ErrorEWMA.Add(float64(pending.Size) / float64(r.Received))
				}
				log.T.Ln("first", pending.First.UnixNano(), "last",
					pending.Last.UnixNano(), r.Ping.Nanoseconds())
				tot := pending.Last.UnixNano() - pending.First.UnixNano()
				pn := r.Ping
				div := float64(pn) / float64(tot)
				log.T.Ln("div", div, "tot", tot)
				d.PingDivergence.Add(div)
				par := float64(d.Parity.Load())
				pv := par * d.PingDivergence.Value() *
					(1 + d.ErrorEWMA.Value())
				log.T.Ln("pv", par, "*", d.PingDivergence.Value(), "*",
					1+d.ErrorEWMA.Value(), "=", pv)
				d.Parity.Store(uint32(byte(pv)))
				log.T.Ln("ack: processed for",
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
		// d.Ready.Signal()
	case OnionMagic:
		o := c.(*Onion)
		d.Duplex.Receiver.Send(o.Bytes)
		go func() {
			// Send out the acknowledgement.
			d.SendAck(rxr)
			// log.T.Ln("ack: sent")
		}()
		// d.Ready.Signal()
		// <-d.Ready
	}
	// d.Ready.Signal()
	log.D.Ln("ready!")
	// <-d.Ready
}

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
