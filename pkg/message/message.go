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
	// retransmit avoidance.
	DataShards, ParityShards slice.Size16
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
	Seen []pub.Print
}

const PacketDataMinSize = pub.PrintLen + slice.Uint16Len*4 + nonce.Size + sig.Len

type EP struct {
	To      *pub.Key
	From    *prv.Key
	Blk     cipher.Block
	DShards int
	PShards int
	Seq     int
	Tot     int
	Data    []byte
	Seen    []pub.Print
}

// Encode creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func Encode(ep EP) (pkt []byte, e error) {
	f := &Packet{
		To:           ep.To.ToBytes().Fingerprint(),
		DataShards:   slice.NewUint16(),
		ParityShards: slice.NewUint16(),
		Seq:          slice.NewUint16(),
		Tot:          slice.NewUint16(),
		Nonce:        nonce.Get(),
		Seen:         ep.Seen,
	}
	slice.EncodeUint16(f.DataShards, ep.DShards)
	slice.EncodeUint16(f.ParityShards, ep.PShards)
	slice.EncodeUint16(f.Seq, ep.Seq)
	slice.EncodeUint16(f.Tot, ep.Tot)
	SeenCount := []byte{byte(len(ep.Seen))}
	payloadLen := slice.NewUint32()
	slice.EncodeUint32(payloadLen, len(ep.Data))
	// Encrypt the payload
	ciph.Encipher(ep.Blk, f.Nonce, ep.Data)
	f.Payload = ep.Data
	var seenBytes []byte
	for i := range f.Seen {
		seenBytes = append(seenBytes, f.Seen[i][:]...)
	}
	pkt = slice.Concatenate(
		f.To[:],
		f.DataShards,
		f.ParityShards,
		f.Seq[:],
		f.Tot,
		f.Nonce[:],
		payloadLen,
		f.Payload,
		SeenCount,
		seenBytes,
	)
	// Sign the packet.
	var s sig.Bytes
	if s, e = sig.Sign(ep.From, sha256.Single(pkt)); !check(e) {
		// Signature space is pre-allocated so we copy it.
		pkt = append(pkt, s...)
	}
	return
}

// Decode a packet and return the Packet with encrypted payload and signer's
// public key.
func Decode(pkt []byte) (f *Packet, p *pub.Key, e error) {
	pktLen := len(pkt)
	if pktLen < PacketDataMinSize {
		// If this isn't checked the slice operations later can
		// hit bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			PacketDataMinSize, pktLen)
		log.E.Ln(e)
		return
	}
	// split off the signature and recover the public key
	sigStart := pktLen - sig.Len
	var d []byte
	var s sig.Bytes
	d, s = pkt[:sigStart], pkt[sigStart:]
	if p, e = s.Recover(sha256.Single(d)); check(e) {
		e = fmt.Errorf("error: '%s': packet checksum failed", e.Error())
	}
	f = &Packet{}
	f.To, d = slice.Cut(d, pub.PrintLen)
	f.DataShards, d = slice.Cut(d, slice.Uint16Len)
	f.ParityShards, d = slice.Cut(d, slice.Uint16Len)
	f.Seq, d = slice.Cut(d, slice.Uint16Len)
	f.Tot, d = slice.Cut(d, slice.Uint16Len)
	f.Nonce, d = slice.Cut(d, nonce.Size)
	var pl slice.Size32
	pl, d = slice.Cut(d, slice.Uint32Len)
	f.Payload, d = slice.Cut(d, slice.DecodeUint32(pl))
	var sc byte
	sc, d = d[0], d[1:]
	var sn []byte
	f.Seen = make([]pub.Print, sc)
	for i := 0; i < int(sc); i++ {
		sn, d = slice.Cut(d, pub.PrintLen)
		copy(f.Seen[i][:], sn)
	}
	return
}
