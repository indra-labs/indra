package engine

import (
	"context"
	"sync"
	"time"
	
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
	// many packets led to a receive against the ratio of DataSent
	// / ParitySent * PingDivergence.
	Parity byte
	// DataSent is the amount of payload bytes sent.
	DataSent int
	// ParitySent is the amount of bytes that were needed to successfully
	// transmit, including packet overhead. This is collected from
	// acknowledgements.
	ParitySent int
	// PingDivergence represents the proportion of time between start of send
	// and receiving acknowledgement, versus the ping RTT being actively
	// measured concurrently. Shorter time can reduce redundancy, longer time
	// needs to increase it. Combined with DataSent / ParitySent this guides the
	// error correction parameter for a given transmission that minimises
	// latency. Value is a binary, fixed point value with 1<<32 as "1".
	PingDivergence  int
	PendingInbound  []*RxRecord
	PendingOutbound []*TxRecord
	*DuplexByteChan
	*Conn
	*crypto.KeySet
	sync.Mutex
}

func NewDispatcher(l *Conn, ctx context.Context,
	ks *crypto.KeySet) (d *Dispatcher) {
	
	d = &Dispatcher{Conn: l, KeySet: ks, Parity: DefaultStartingParity,
		DuplexByteChan: NewDuplexByteChan(ConnBufs)}
	go func() {
		for {
			select {
			case m := <-l.Recv:
				log.D.S("received from conn to dispatcher", m.ToBytes())
			case m := <-d.DuplexByteChan.Send:
				// Data received for sending through the Conn.
				_ = m
			case <-ctx.Done():
				return
			}
		}
	}()
	return
}

func (d *Dispatcher) Split(pp *PacketParams) (pkts Packets,
	packets [][]byte, e error) {
	
	return SplitToPackets(pp, d.MTU, d.KeySet)
}

func (d *Dispatcher) Join(packets Packets) (pkts Packets, msg []byte, e error) {
	return JoinPackets(packets)
}
