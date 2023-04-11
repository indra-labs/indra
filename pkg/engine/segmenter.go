package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

type TxRecord struct {
	ID nonce.ID
	// First is when the first packet arrived.
	First time.Time
	// Last is when the last packet arrived that led to successful
	// reconstruction.
	Last time.Time
	// Minimum is the size of the data packet.
	Minimum int
	// Required is the number of bytes that were collected for the eventual
	// receiving.
	Required int
}

type Segmenter struct {
	*Listener
	*crypto.KeySet
	// MinReq is the count of packets for successful transmission.
	MinReq int
	// ActualReq is the amount of bytes that were needed to successfully
	// transmit.
	ActualReq int
	// PingDivergence represents the proportion of time between start of send
	// and receiving acknowledgement, versus the ping RTT being actively
	// measured. Shorter time can reduce redundancy, longer time needs to
	// increase it. Combined with MinReq / ActualReq this guides the error
	// correction parameter for a given transmission that minimises latency.
	// Value is a binary fixed point value with 1<<1024 as "1".
	PingDivergence int
	PendingAcks    []*TxRecord
}

func NewSegmenter(l *Listener, ks *crypto.KeySet) (s *Segmenter) {
	s = &Segmenter{Listener: l, KeySet: ks}
	go func() {
		for {
			select {
			case <-l.Context.Done():
				return
			case c := <-l.newConns:
				_ = c
				// here start workers for connections.
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
