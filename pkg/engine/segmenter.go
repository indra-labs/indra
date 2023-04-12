package engine

import (
	"sync"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
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
// in a message received acknowledgement. The exported fields go into an
// acknowledgement message.
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

// Segmenter is a message splitter/joiner and error correction adjustment system
// that aims to minimise message latency by trading it for bandwidth especially
// to cope with radio connections.
type Segmenter struct {
	*Listener
	*crypto.KeySet
	// DataSent is the amount of actual bytes sent.
	DataSent int
	// ParitySent is the amount of bytes that were needed to successfully
	// transmit, including packet overhead.
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
	sync.Mutex
}

func NewSegmenter(l *Listener, ks *crypto.KeySet) (s *Segmenter) {
	s = &Segmenter{Listener: l, KeySet: ks}
	go func() {
		for {
			select {
			case <-l.Context.Done():
				return
			}
		}
	}()
	return
}

func (s *Segmenter) Split(pp *PacketParams) (pkts Packets,
	packets [][]byte, e error) {
	
	return SplitToPackets(pp, s.Listener.MTU, s.KeySet)
}

func (s *Segmenter) Join(packets Packets) (pkts Packets, msg []byte, e error) {
	return JoinPackets(packets)
}
