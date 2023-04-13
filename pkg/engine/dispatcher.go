package engine

import (
	"context"
	"sync"
	"time"
	
	"github.com/VividCortex/ewma"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
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
	Received int
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
// DataSent / ParitySent provides the ratio of redundancy the channel is using.
// ParitySent is not from the parameters at send but from acknowledgements of
// how much data was received before a message was reconstructed. Thus, it is
// used in combination with the PingDivergence to recompute the Parity parameter
// used for adjusting error correction redundancy as each message is decoded.
type Dispatcher struct {
	// Parity is the parity parameter to use in packet segments, this value
	// should be adjusted up and down in proportion with collected data on how
	// many packets led to receipt against the ratio of DataSent / ParitySent *
	// PingDivergence.
	Parity atomic.Uint32
	// DataSent is the amount of payload bytes sent.
	DataSent int
	// ParitySent is the amount of bytes that were needed to successfully
	// transmit, including packet overhead. This is collected from
	// acknowledgements.
	ParitySent int
	// PingDivergence represents the proportion of time between start of send
	// and receiving acknowledgement, versus the ping RTT being actively
	// measured concurrently. Shorter/equal time means it can reduce redundancy,
	// longer time needs to increase it.
	//
	// Combined with DataSent / ParitySent this guides the error correction
	// parameter for a given transmission that minimises latency. Onion routing
	// necessarily amplifies any latency so making a transmission get across
	// before/without retransmits is as good as the path can provide.
	PingDivergence int
	Duplex         *DuplexByteChan
	*Conn
	*crypto.Prv
	*crypto.KeySet
	sync.Mutex
	Ping            ewma.MovingAverage
	PendingInbound  []*RxRecord
	PendingOutbound []*TxRecord
	Partials        map[nonce.ID]Packets
}

const TimeoutPingCount = 10

func NewDispatcher(l *Conn, ctx context.Context,
	ks *crypto.KeySet) (d *Dispatcher) {
	
	d = &Dispatcher{Conn: l, KeySet: ks,
		Duplex: NewDuplexByteChan(ConnBufs), Ping: ewma.NewMovingAverage()}
	d.Parity.Store(DefaultStartingParity)
	ps := ping.NewPingService(l.Host)
	pings := ps.Ping(ctx, l.Conn.RemotePeer())
	garbage := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-garbage.C:
				// remove successful receives after they are finished or all
				// pieces arrived.
				go func() {
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
									break
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
					for i := range rxr {
						// send the RxRecord to the peer.
						_ = rxr[i]
					}
				}()
			case p := <-pings:
				d.Lock()
				d.Ping.Add(float64(p.RTT))
				d.Unlock()
			case m := <-l.Recv:
				log.D.S("received from conn to dispatcher", m.ToBytes())
				// Packet received, decrypt, gather and send acks back and
				// reconstructed messages to the Dispatcher.Recv channel.
				from, to, iv, e := GetKeysFromPacket(m)
				if fails(e) {
					continue
				}
				if !crypto.Match(to, crypto.DerivePub(d.Prv).ToBytes()) {
					// This connection should only receive messages with cloaked
					// keys matching our private key.
					continue
				}
				d.Lock()
				var p *Packet
				if p, e = DecodePacket(m, from, d.Prv, iv); fails(e) {
					d.Unlock()
					continue
				}
				d.Unlock()
				// if the message only one packet we can send it out now:
				if int(p.Length) == len(p.Data) {
					log.D.Ln("forwarding single packet message")
					d.Recv <- m
					continue
				}
				// Find collection of existing fragments matching the message ID
				// or make a new one and add this packet to it for later
				// assembly.
				log.D.Ln("collating packet")
				d.Lock()
				d.Partials[p.ID] = append(d.Partials[p.ID], p)
				segCount := int(p.Length) / d.Conn.MTU
				mod := int(p.Length) % d.Conn.MTU
				if mod > 0 {
					segCount++
				}
				if len(d.Partials[p.ID]) >= segCount {
					// Enough to attempt reconstruction:
					var msg []byte
					if d.Partials[p.ID], msg, e = JoinPackets(d.Partials[p.ID]); fails(e) {
						continue
					}
					log.D.Ln("message reconstructed; dispatching...")
					d.Recv <- msg
				}
				d.Unlock()
			case m := <-d.Duplex.Send:
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
					To:     l.RemoteKey,
					Parity: int(d.Parity.Load()),
					Length: m.Len(),
					Data:   m,
				}
				packets, e := SplitToPackets(pp, l.MTU, ks)
				if fails(e) {
					continue
				}
				for i := range packets {
					l.DuplexByteChan.Send <- packets[i]
				}
				txr.Last = time.Now()
				txr.Ping = time.Duration(d.Ping.Value())
				d.Lock()
				d.PendingOutbound = append(d.PendingOutbound, txr)
				d.Unlock()
				log.D.Ln("message dispatched")
			case <-ctx.Done():
				return
			}
		}
	}()
	return
}
