package message

import (
	"crypto/cipher"
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/sig"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Packet is the standard format for an encrypted, possibly segmented message
// container with parameters for Reed Solomon Forward Error Correction and
// contains previously seen cipher keys so the correspondent can free them.
type Packet struct {
	// To is the fingerprint of the pubkey used in the ECDH key exchange, 12
	// bytes long.
	To pub.Print
	// DataShards and ParityShards are reed solomon parameters for
	// retransmit avoidance. These are a ratio and
	// Tot / ParityShards * DataShards will define how many segments of the
	// total are parity shards.
	DataShards, ParityShards byte
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq slice.Size16
	// Tot is the number of segments in the batch
	Tot slice.Size16
	// Nonce is the IV for the encryption on the Payload. 16 bytes.
	Nonce nonce.IV
	// Payload is the encrypted message.
	Payload []byte
	// Seen is the SHA256 truncated hashes of previous received encryption
	// public keys to indicate they won't be reused and can be discarded.
	// The binary encoding allows for 256 of these
	Seen []pub.Print
}

const PacketOverhead = pub.PrintLen + slice.Uint16Len*4 + nonce.Size + sig.Len

type EP struct {
	To         *pub.Key
	From       *prv.Key
	Blk        cipher.Block
	DShards    int
	PShards    int
	Seq        int
	Tot        int
	Data       []byte
	Seen       []pub.Print
	PayloadLen int
}

func (ep EP) GetOverhead() int {
	return PacketOverhead + len(ep.Seen)*pub.PrintLen
}

// Encode creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func Encode(ep EP) (pkt []byte, e error) {
	f := &Packet{
		To:           ep.To.ToBytes().Fingerprint(),
		DataShards:   byte(ep.DShards),
		ParityShards: byte(ep.PShards),
		Seq:          slice.NewUint16(),
		Tot:          slice.NewUint16(),
		Nonce:        nonce.Get(),
		Seen:         ep.Seen,
	}
	slice.EncodeUint16(f.Seq, ep.Seq)
	slice.EncodeUint16(f.Tot, ep.Tot)
	SeenCount := []byte{byte(len(ep.Seen))}
	// We are not supporting packets longer than 64kb as this is the largest
	// supported by current network devices for UDP packets. Larger messages
	// must be split. The length of 16 here also is to ensure that the
	// actual payload starts on a 16 byte boundary to be optimal for the
	// AES-CTR encryption, the preceding data total size is 16 bytes.
	payloadLen := slice.NewUint16()
	slice.EncodeUint16(payloadLen, ep.PayloadLen)
	// Encrypt the payload
	ciph.Encipher(ep.Blk, f.Nonce, ep.Data)
	f.Payload = ep.Data
	var seenBytes []byte
	for i := range f.Seen {
		seenBytes = append(seenBytes, f.Seen[i][:]...)
	}
	pkt = slice.Concatenate(
		f.To[:],                // 8 bytes   \
		[]byte{f.DataShards},   // 1 byte     |
		[]byte{f.ParityShards}, // 1 byte     |
		f.Seq[:],               // 2 bytes     > 32 bytes
		f.Tot,                  // 2 bytes    |
		payloadLen,             // 2 byte     |
		f.Nonce[:],             // 16 bytes  /
		f.Payload,              // payload starts on 32 byte boundary
		SeenCount,
		seenBytes,
	)
	// Sign the packet.
	var s sig.Bytes
	if s, e = sig.Sign(ep.From, sha256.Single(pkt)); !check(e) {
		pkt = append(pkt, s...)
	}
	return
}

// Decode a packet and return the Packet with encrypted payload and signer's
// public key.
func Decode(pkt []byte) (f *Packet, p *pub.Key, e error) {
	const (
		u16l = slice.Uint16Len
		prl  = pub.PrintLen
	)
	pktLen := len(pkt)
	if pktLen < PacketOverhead {
		// If this isn't checked the slice operations later can
		// hit bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			PacketOverhead, pktLen)
		log.E.Ln(e)
		return
	}
	// split off the signature and recover the public key
	sigStart := pktLen - sig.Len
	var data []byte
	var s sig.Bytes
	data, s = pkt[:sigStart], pkt[sigStart:]
	if p, e = s.Recover(sha256.Single(data)); check(e) {
		e = fmt.Errorf("error: '%s': packet checksum failed", e.Error())
	}
	f = &Packet{}
	f.To, data = slice.Cut(data, prl)
	f.DataShards, data = data[0], data[1:]
	f.ParityShards, data = data[0], data[1:]
	f.Seq, data = slice.Cut(data, u16l)
	f.Tot, data = slice.Cut(data, u16l)
	var payloadLength slice.Size16
	payloadLength, data = slice.Cut(data, u16l)
	f.Nonce, data = slice.Cut(data, nonce.Size)
	f.Payload, data = slice.Cut(data, slice.DecodeUint16(payloadLength))
	var sc byte
	sc, data = data[0], data[1:]
	var sn []byte
	f.Seen = make([]pub.Print, sc)
	for i := 0; i < int(sc); i++ {
		sn, data = slice.Cut(data, pub.PrintLen)
		copy(f.Seen[i][:], sn)
	}
	return
}
