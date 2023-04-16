package engine

import (
	"context"
	"math/big"
	"sync"
	"time"
	
	"github.com/VividCortex/ewma"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	// DefaultStartingParity is set to 64, or 25%
	DefaultStartingParity = 64
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
	sha256.Hash
	// First is when the first packet was received.
	First time.Time
	// Last is when the last packet was received. A longer time than the current
	// ping RTT after First indicates retransmits.
	Last time.Time
	// Size of the message as found in the packet headers.
	Size int
	// Received is the number of bytes received upon reconstruction, including
	// packet overhead.
	Received uint64
	// Ping is the average ping RTT on the connection calculated at each packet
	// receive, used with the total message transmit time to estimate an
	// adjustment in the parity shards to be used in sending on this connection.
	Ping time.Duration
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
	Duplex          *DuplexByteChan
	PendingInbound  []*RxRecord
	PendingOutbound []*TxRecord
	Partials        map[nonce.ID]Packets
	OldPrv          *crypto.Prv
	Prv             *crypto.Prv
	*crypto.KeySet
	Conn *Conn
	sync.Mutex
}

const TimeoutPingCount = 10

func NewDispatcher(l *Conn, key *crypto.Prv, ctx context.Context,
	ks *crypto.KeySet) (d *Dispatcher) {
	
	d = &Dispatcher{
		Prv:            key,
		OldPrv:         key,
		Conn:           l,
		KeySet:         ks,
		Duplex:         NewDuplexByteChan(ConnBufs),
		Ping:           ewma.NewMovingAverage(),
		PingDivergence: ewma.NewMovingAverage(),
		ErrorEWMA:      ewma.NewMovingAverage(),
	}
	d.Parity.Store(DefaultStartingParity)
	ps := ping.NewPingService(l.Host)
	pings := ps.Ping(ctx, l.Conn.RemotePeer())
	garbageTicker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-garbageTicker.C:
				go d.RunGC()
			case p := <-pings:
				go d.HandlePing(p)
			case m := <-l.Transport.Receive():
				go d.RecvFromConn(m)
			case m := <-d.Duplex.Sender.Receive():
				go d.SendToConn(m)
			case <-ctx.Done():
				return
			}
		}
	}()
	return
}

func (d *Dispatcher) RunGC() {
	// remove successful receives after all pieces arrived. Successful
	// transmissions before the timeout will already be cleared from
	// confirmation by the acknowledgment.
	var rxr []*RxRecord
	d.Lock()
	for dpi, dp := range d.Partials {
		// Find the oldest and newest.
		oldest := time.Now()
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
			
			// after 10 pings of time elapse since last received we
			// consider the transmission a failure, send back the
			// failed RxRecord and delete the map entry.
			var tmp []*RxRecord
			for i := range d.PendingInbound {
				if d.PendingInbound[i].ID == dp[0].ID {
					rxr = append(rxr, d.PendingInbound[i])
				} else {
					tmp = append(tmp, d.PendingInbound[i])
				}
			}
			delete(d.Partials, dpi)
			d.PendingInbound = tmp
		}
	}
	d.Unlock()
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
	d.Lock()
	d.Ping.Add(float64(p.RTT))
	d.Unlock()
}

func (d *Dispatcher) RecvFromConn(m slice.Bytes) {
	log.D.S("received from conn to dispatcher", m.ToBytes())
	// Packet received, decrypt, gather and send acks back and reconstructed
	// messages to the Dispatcher.RecvFromConn channel.
	from, to, iv, e := GetKeysFromPacket(m)
	if fails(e) {
		return
	}
	d.Lock()
	// This connection should only receive messages with cloaked keys
	// matching our private key of the connection.
	prv := d.Prv
	if !crypto.Match(to, crypto.DerivePub(d.Prv).ToBytes()) {
		prv = d.OldPrv
	}
	if !crypto.Match(to, crypto.DerivePub(d.OldPrv).ToBytes()) {
		d.Unlock()
		return
	}
	var p *Packet
	if p, e = DecodePacket(m, from, prv, iv); fails(e) {
		d.Unlock()
		return
	}
	var rxr *RxRecord
	var packets Packets
	if rxr, packets = d.GetRxRecordAndPartials(p.ID); rxr != nil {
		rxr.Received += uint64(len(m))
		rxr.Last = time.Now()
		packets = append(packets, p)
	} else {
		rxr = &RxRecord{
			ID:       p.ID,
			First:    time.Now(),
			Last:     time.Now(),
			Size:     int(p.Length),
			Received: uint64(len(m)),
			Ping:     time.Duration(d.Ping.Value()),
		}
		d.PendingInbound = append(d.PendingInbound, rxr)
		packets = Packets{p}
		d.Partials[p.ID] = packets
	}
	d.Unlock()
	// if the message is only one packet we can hand it on to the receiving
	// channel now:
	if int(p.Length) == len(p.Data) {
		log.D.Ln("forwarding single packet message")
		d.Handle(m, rxr)
		return
	}
	// Find collection of existing fragments matching the message ID or make a
	// new one and add this packet to it for later assembly.
	log.D.Ln("collating packet")
	d.Lock()
	d.TotalReceived = d.TotalReceived.Add(d.TotalReceived,
		big.NewInt(int64(len(m))))
	d.Partials[p.ID] = append(d.Partials[p.ID], p)
	d.Unlock()
	segCount := int(p.Length) / d.Conn.GetMTU()
	mod := int(p.Length) % d.Conn.GetMTU()
	if mod > 0 {
		segCount++
	}
	d.Lock()
	defer d.Unlock()
	if len(d.Partials[p.ID]) >= segCount {
		// Enough to attempt reconstruction:
		var msg []byte
		if d.Partials[p.ID], msg, e = JoinPackets(d.Partials[p.ID]); fails(e) {
			log.D.Ln("failed to join packets")
			return
		}
		log.D.Ln("message reconstructed; dispatching...")
		d.DataReceived = d.DataReceived.Add(d.DataReceived,
			big.NewInt(int64(len(msg))))
		for _, v := range d.Partials[p.ID] {
			d.TotalReceived = d.TotalReceived.Add(
				d.TotalReceived,
				big.NewInt(int64(len(v.Data)+v.GetOverhead())),
			)
		}
		// Sender the message on to the receiving channel.
		d.Handle(msg, rxr)
	}
}

func (d *Dispatcher) SendAck(rxr *RxRecord) {
	ack := &Acknowledge{rxr}
	s := splice.New(ack.Len())
	_ = ack.Encode(s)
	d.Duplex.Send(s.GetAll())
	// Remove Rx from pending.
	d.Lock()
	var tmp []*RxRecord
	for _, v := range d.PendingInbound {
		if rxr.ID != v.ID {
			tmp = append(tmp, v)
		}
	}
	d.PendingInbound = tmp
	d.Unlock()
}

func (d *Dispatcher) SendToConn(m slice.Bytes) {
	log.D.S("message dispatching to conn", m.ToBytes())
	// Data received for sending through the Conn.
	id := nonce.NewID()
	hash := sha256.Single(m)
	txr := &TxRecord{
		ID:    id,
		Hash:  hash,
		First: time.Now(),
		Size:  len(m),
	}
	pp := &PacketParams{
		ID:     id,
		To:     d.Conn.GetRemoteKey(),
		Parity: int(d.Parity.Load()),
		Length: m.Len(),
		Data:   m,
	}
	mtu := d.Conn.GetMTU()
	packets, e := SplitToPackets(pp, mtu, d.KeySet)
	if fails(e) {
		return
	}
	// Shuffle. This is both for anonymity and improving the chances of most
	// error bursts to not cut through a whole segment (they are grouped by 256
	// for RS FEC).
	cryptorand.Shuffle(len(packets), func(i, j int) {
		packets[i], packets[j] = packets[j], packets[i]
	})
	// Sender them out!
	sendChan := d.Conn.GetSend()
	for i := range packets {
		sendChan.Send(packets[i])
	}
	txr.Last = time.Now()
	txr.Ping = time.Duration(d.Ping.Value())
	d.Lock()
	for _, v := range packets {
		d.TotalSent = d.TotalSent.Add(d.TotalSent,
			big.NewInt(int64(len(v))))
	}
	d.PendingOutbound = append(d.PendingOutbound, txr)
	d.Unlock()
	log.D.Ln("message dispatched")
}

func (d *Dispatcher) Handle(m slice.Bytes, rxr *RxRecord) {
	// Sender out the acknowledgement.
	d.SendAck(rxr)
	s := splice.NewFrom(m)
	c := Recognise(s)
	if c == nil {
		return
	}
	magic := c.Magic()
	log.D.Ln("decoding message with magic", magic)
	var e error
	if e = c.Decode(s); fails(e) {
		return
	}
	switch magic {
	case InitRekeyMagic:
		o := c.(*InitRekey)
		log.D.S("key change initiate", o)
		d.Conn.SetRemoteKey(o.NewPubkey)
		var prv *crypto.Prv
		if prv, e = crypto.GeneratePrvKey(); fails(e) {
			return
		}
		d.Lock()
		d.OldPrv = d.Prv
		d.Prv = prv
		d.Unlock()
		// Sender a reply:
		rpl := RekeyReply{NewPubkey: crypto.DerivePub(prv)}
		reply := splice.New(rpl.Len())
		if e = rpl.Encode(reply); fails(e) {
			return
		}
		d.Duplex.Send(reply.GetAll())
	case RekeyReplyMagic:
		o := c.(*RekeyReply)
		log.D.S("key change reply", o)
		d.Conn.SetRemoteKey(o.NewPubkey)
	case AcknowledgeMagic:
		o := c.(*Acknowledge)
		log.D.S("acknowledgement", o)
		r := o.RxRecord
		d.Lock()
		var tmp []*TxRecord
		for _, pending := range d.PendingOutbound {
			if pending.ID == r.ID {
				// Accounting of successful send.
				if r.Hash == pending.Hash {
					d.DataSent = d.DataSent.Add(d.DataSent,
						big.NewInt(int64(pending.Size)))
				}
				d.ErrorEWMA.Add(float64(r.Received) / float64(r.Size))
				tot := r.Last.Sub(pending.First) * 2
				div := float64(tot) / float64(r.Ping)
				d.PingDivergence.Add(div)
				par := float64(d.Parity.Load())
				d.Parity.Store(uint32(byte(
					par * d.PingDivergence.Value() *
						d.ErrorEWMA.Value())))
			} else {
				tmp = append(tmp, pending)
			}
		}
		// Entry is now deleted and processed.
		d.PendingOutbound = tmp
		d.Unlock()
	case MungedMagic:
		o := c.(*Munged)
		log.D.S("mung!", o)
		
	}
}

func (d *Dispatcher) GetRxRecordAndPartials(id nonce.ID) (rxr *RxRecord,
	packets Packets) {
	
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
