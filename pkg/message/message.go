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
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq uint16
	// Length is the number of segments in the batch
	Length uint32
	// Parity is the ratio of redundancy. In each 256 segment
	Parity byte
	// Nonce is the IV for the encryption on the Payload. 16 bytes.
	Nonce nonce.IV
	// Payload is the encrypted message.
	Data []byte
	// Seen is the SHA256 truncated hashes of previous received encryption
	// public keys to indicate they won't be reused and can be discarded.
	// The binary encoding allows for 256 of these
	Seen []pub.Print
}

// GetOverhead returns the packet frame overhead given the settings found in the
// packet.
func (p *Packet) GetOverhead() int {
	return Overhead + len(p.Seen)*pub.PrintLen
}

// Decipher reverses the current state of encryption of the packet.
func (p *Packet) Decipher(blk cipher.Block) *Packet {
	ciph.Encipher(blk, p.Nonce, p.Data)
	return p
}

// Overhead is the base overhead on a packet, use GetOverhead to add any extra
// as found in a Packet.
const Overhead = pub.PrintLen + 1 + 2 + slice.Uint16Len*3 + nonce.Size +
	sig.Len + 4

// Packets is a slice of pointers to packets.
type Packets []*Packet

// sort.Interface implementation.

func (p Packets) Len() int           { return len(p) }
func (p Packets) Less(i, j int) bool { return p[i].Seq < p[j].Seq }
func (p Packets) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// EP defines the parameters for creating a (split) packet given a set of keys,
// cipher, and data. To, From, Blk and Data are required, Parity is optional,
// set it to define a level of Reed Solomon redundancy on the split packets.
// Seen should be populated to send a signal to the other side of keys that have
// been seen at time of constructing this packet that can now be discarded as
// they will not be used to generate a cipher again.
type EP struct {
	To     *pub.Key
	From   *prv.Key
	Blk    cipher.Block
	Parity int
	Seq    int
	Length int
	Data   []byte
	Seen   []pub.Print
	Pad    int
}

// GetOverhead returns the amount of the message that will not be part of the
// payload.
func (ep EP) GetOverhead() int {
	return Overhead + len(ep.Seen)*pub.PrintLen
}

// Encode creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func Encode(ep EP) (pkt []byte, e error) {
	f := &Packet{
		To:    ep.To.ToBytes().Fingerprint(),
		Nonce: nonce.Get(),
		Seen:  ep.Seen,
	}
	parity := []byte{byte(ep.Parity)}
	Seq := slice.NewUint16()
	Tot := slice.NewUint32()
	slice.EncodeUint16(Seq, ep.Seq)
	slice.EncodeUint32(Tot, ep.Length)
	SeenCount := []byte{byte(len(ep.Seen))}
	// Encrypt the payload
	ciph.Encipher(ep.Blk, f.Nonce, ep.Data)
	f.Data = ep.Data
	var seenBytes []byte
	for i := range f.Seen {
		seenBytes = append(seenBytes, f.Seen[i][:]...)
	}
	// Concatenate the message pieces together into a single byte slice.
	pkt = slice.Concatenate(
		f.To[:],    // 6 bytes  \
		Seq,        // 2 bytes   |
		Tot,        // 4 bytes   |
		parity,     // 1 byte    |
		SeenCount,  // 1 byte    |
		f.Nonce[:], // 16 bytes  /
		f.Data,     // payload starts on 32 byte boundary
		seenBytes,
	)
	// Sign the packet.
	var s sig.Bytes
	hash := sha256.Single(pkt)
	pkt = append(pkt, hash[:4]...)
	if s, e = sig.Sign(ep.From, hash); !check(e) {
		pkt = append(pkt, s...)
	}
	return
}

// Decode a packet and return the Packet with encrypted payload and signer's
// public key.
func Decode(d []byte) (f *Packet, p *pub.Key, e error) {
	const (
		u16l = slice.Uint16Len
		u32l = slice.Uint32Len
		prl  = pub.PrintLen
	)
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can
		// hit bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	// split off the signature and recover the public key
	sigStart := pktLen - sig.Len
	var s sig.Bytes
	var data, chek []byte
	s, data = d[sigStart:], d[:sigStart]
	checkStart := len(data) - 4
	chek, data = data[checkStart:], data[:checkStart]
	hash := sha256.Single(data)
	if string(chek) != string(hash[:4]) {
		e = fmt.Errorf("check failed: got '%v', expected '%v'",
			chek, hash[:4])
		return
	}
	if p, e = s.Recover(hash); check(e) {
		return
	}
	// log.I.Ln("pktLen", pktLen, "sigStart", sigStart)
	f = &Packet{}
	f.To, data = slice.Cut(data, prl)
	var seq, tot slice.Size16
	seq, data = slice.Cut(data, u16l)
	f.Seq = uint16(slice.DecodeUint16(seq))
	tot, data = slice.Cut(data, u32l)
	f.Length = uint32(slice.DecodeUint32(tot))
	f.Parity, data = data[0], data[1:]
	var sc byte
	sc, data = data[0], data[1:]
	f.Nonce, data = slice.Cut(data, nonce.Size)
	pl := len(data) - int(sc)
	// log.I.Ln(f.Seq, pl)
	f.Data, data = slice.Cut(data, pl)
	// trim the padding
	data = data[:len(data)-int(sc)*pub.PrintLen]
	var sn []byte
	f.Seen = make([]pub.Print, sc)
	for i := 0; i < int(sc); i++ {
		sn, data = slice.Cut(data, pub.PrintLen)
		copy(f.Seen[i][:], sn)
	}
	return
}
