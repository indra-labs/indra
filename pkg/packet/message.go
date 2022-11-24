// Package packet provides a standard message binary serialised data format and
// message segmentation scheme which includes address.Sender cloaked public
// key and address.Receiver private keys for generating a shared cipher and applying
// to messages/message segments.
package packet

import (
	"crypto/cipher"
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/blake3"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/sig"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/nonce"
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
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq uint16
	// Length is the number of segments in the batch
	Length uint32
	// Parity is the ratio of redundancy. In each 256 segment
	Parity byte
	// Return is the return address to use for a reply. Even if only one
	// segment of a segmented message is received the key can be used as a
	// recipient encryption key half.
	Return pub.Bytes
	// Data is the message.
	Data []byte
}

// GetOverhead returns the packet frame overhead given the settings found in the
// packet.
func (p *Packet) GetOverhead() int {
	return Overhead
}

// Overhead is the base overhead on a packet, use GetOverhead to add any extra
// as found in a Packet.
const Overhead = nonce.IVLen + address.Len + slice.Uint16Len +
	slice.Uint32Len + 1 + 4 + sig.Len

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
	To     *address.Sender
	From   *signer.KeySet
	Parity int
	Seq    int
	Length int
	Data   []byte
}

// GetOverhead returns the amount of the message that will not be part of the
// payload.
func (ep EP) GetOverhead() int {
	return Overhead
}

// Encode creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func Encode(ep EP) (pkt []byte, e error) {
	var blk cipher.Block
	if blk, e = ciph.GetBlock(ep.From.Next(), ep.To.Key); check(e) {
		return
	}
	nonc := nonce.New()
	var to address.Cloaked
	to, e = ep.To.GetCloak()
	parity := []byte{byte(ep.Parity)}
	Seq := slice.NewUint16()
	Tot := slice.NewUint32()
	slice.EncodeUint16(Seq, ep.Seq)
	slice.EncodeUint32(Tot, ep.Length)
	// Concatenate the message pieces together into a single byte slice.
	pkt = slice.Cat(
		// f.Nonce[:],    // 16 bytes \
		// f.To[:],       // 8 bytes   |
		make([]byte,
			nonce.IVLen+address.Len),
		Seq,    // 2 bytes
		Tot,    // 4 bytes
		parity, // 1 byte
		ep.Data,
	)
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, pkt)
	// put nonce and recipient print in place.
	copy(pkt, append(nonc, to...))
	// Sign the packet.
	var s sig.Bytes
	hash := blake3.Single(pkt)
	if s, e = sig.Sign(ep.From.Current(), hash); check(e) {
		return
	}
	pkt = append(pkt, hash[:4]...)
	pkt = append(pkt, s...)
	return
}

// GetKeys returns the To field of the message in order, checks the packet
// checksum and recovers the public key signing it.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be processed with ciph.Encipher (sans signature)
// using the block cipher thus created from the shared secret, and the Decode
// function will then decode a Packet.
func GetKeys(d []byte) (to address.Cloaked, from *pub.Key, e error) {
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can
		// hit bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	to = d[nonce.IVLen : nonce.IVLen+address.Len]
	// split off the signature and recover the public key
	sigStart := pktLen - sig.Len
	checkStart := sigStart - 4
	var s sig.Bytes
	var chek []byte
	s = d[sigStart:]
	chek = d[checkStart:sigStart]
	hash := blake3.Single(d[:checkStart])
	if string(chek) != string(hash[:4]) {
		e = fmt.Errorf("check failed: got '%v', expected '%v'",
			chek, hash[:4])
		return
	}
	if from, e = s.Recover(hash); check(e) {
		return
	}
	return
}

// Decode a packet and return the Packet with encrypted payload and signer's
// public key. This assumes GetKeys succeeded and the matching private key was
// found.
func Decode(d []byte, from *pub.Key, to *prv.Key) (f *Packet, e error) {
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can
		// hit bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	// Trim off the signature and hash, we already have the key and have
	// validated the checksum.
	sigStart := pktLen - sig.Len - 4
	data := d[:sigStart]
	f = &Packet{}
	nonc := data[:nonce.IVLen]
	var blk cipher.Block
	if blk, e = ciph.GetBlock(to, from); check(e) {
		return
	}
	// This decrypts the rest of the packet, which is encrypted for
	// security.
	ciph.Encipher(blk, nonc, data)
	// Trim off the nonce and recipient fingerprint, which is now encrypted.
	data = data[nonce.IVLen+address.Len:]
	var seq, tot slice.Size16
	seq, data = slice.Cut(data, slice.Uint16Len)
	f.Seq = uint16(slice.DecodeUint16(seq))
	tot, data = slice.Cut(data, slice.Uint32Len)
	f.Length = uint32(slice.DecodeUint32(tot))
	f.Parity, data = data[0], data[1:]
	f.Data = data
	return
}
