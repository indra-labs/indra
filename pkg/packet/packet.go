// Package packet provides a standard message binary serialised data format and
// message segmentation scheme which includes address.Sender cloaked public key
// and address.Receiver private keys for generating a shared cipher and applying
// to messages/message segments.
package packet

import (
	"crypto/cipher"
	"fmt"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Packet is the standard format for an encrypted, possibly segmented message
// container with parameters for Reed Solomon Forward Error Correction.
type Packet struct {
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq uint16
	// Length is the number of segments in the batch
	Length uint32
	// Parity is the ratio of redundancy. In each 256 segment
	Parity byte
	// Data is the message.
	Data []byte
}

// Overhead is the base overhead on a packet, use GetOverhead to add any extra
// as found in a Packet.
const Overhead = 4 + nonce.IVLen + pub.KeyLen + cloak.Len

// Packets is a slice of pointers to packets.
type Packets []*Packet

// sort.Interface implementation.

func (p Packets) Len() int           { return len(p) }
func (p Packets) Less(i, j int) bool { return p[i].Seq < p[j].Seq }
func (p Packets) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Params defines the parameters for creating a (split) packet given a set of
// keys, cipher, and data. To, From, Blk and Data are required, Parity is
// optional, set it to define a level of Reed Solomon redundancy on the split
// packets. Seen should be populated to send a signal to the other side of keys
// that have been seen at time of constructing this packet that can now be
// discarded as they will not be used to generate a cipher again.
type Params struct {
	To     *pub.Key
	From   *prv.Key
	Parity int
	Seq    int
	Length int
	Data   []byte
}

// Encode creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func Encode(p Params) (pkt []byte, e error) {
	var blk cipher.Block
	if blk = ciph.GetBlock(p.From, p.To, "packet encode"); check(e) {
		return
	}
	nonc := nonce.New()
	Seq := slice.NewUint16()
	slice.EncodeUint16(Seq, p.Seq)
	Length := slice.NewUint32()
	slice.EncodeUint32(Length, p.Length)
	pkt = make([]byte, slice.SumLen(Seq, Length, p.Data)+1+Overhead)
	// Append pubkey used for encryption key derivation.
	k := pub.Derive(p.From).ToBytes()
	cloaked := cloak.GetCloak(p.To)
	// Copy nonce, address and key over top of the header.
	c := new(slice.Cursor)
	copy(pkt[c.Inc(4):c.Inc(nonce.IVLen)], nonc[:])
	copy(pkt[*c:c.Inc(cloak.Len)], cloaked[:])
	copy(pkt[*c:c.Inc(pub.KeyLen)], k[:])
	copy(pkt[*c:c.Inc(slice.Uint16Len)], Seq)
	copy(pkt[*c:c.Inc(slice.Uint32Len)], Length)
	pkt[*c] = byte(p.Parity)
	copy(pkt[c.Inc(1):], p.Data)
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, pkt[Overhead:])
	// last but not least, the packet check header, which protects the entire
	// packet.
	checkBytes := sha256.Single(pkt[4:])
	copy(pkt[:4], checkBytes[:4])
	return
}

// GetKeys returns the ToHeaderPub field of the message in order, checks the
// packet checksum and recovers the public key.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be decrypted.
func GetKeys(d []byte) (from *pub.Key, to cloak.PubKey, e error) {
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	// split off the signature and recover the public key
	var k pub.Bytes
	var chek []byte
	c := new(slice.Cursor)
	chek = d[:c.Inc(4)]
	copy(to[:], d[c.Inc(nonce.IVLen):c.Inc(cloak.Len)])
	copy(k[:], d[*c:c.Inc(pub.KeyLen)])
	checkHash := sha256.Single(d[4:])
	if string(chek) != string(checkHash[:4]) {
		e = fmt.Errorf("check failed: got '%v', expected '%v'",
			chek, checkHash[:4])
		return
	}
	if from, e = pub.FromBytes(k[:]); check(e) {
		return
	}
	return
}

// Decode a packet and return the Packet with encrypted payload and signer's
// public key. This assumes GetKeys succeeded and the matching private key was
// found.
func Decode(d []byte, from *pub.Key, to *prv.Key) (p *Packet, e error) {
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	p = &Packet{}
	// copy the nonce
	var nonc nonce.IV
	c := new(slice.Cursor)
	copy(nonc[:], d[c.Inc(4):c.Inc(nonce.IVLen)])
	var blk cipher.Block
	if blk = ciph.GetBlock(to, from, "packet decode"); check(e) {
		return
	}
	// This decrypts the rest of the packet, which is encrypted for security.
	data := d[c.Inc(pub.KeyLen+cloak.Len):]
	ciph.Encipher(blk, nonc, data)
	seq := slice.NewUint16()
	length := slice.NewUint32()
	seq, data = slice.Cut(data, slice.Uint16Len)
	p.Seq = uint16(slice.DecodeUint16(seq))
	length, data = slice.Cut(data, slice.Uint32Len)
	p.Length = uint32(slice.DecodeUint32(length))
	p.Parity, data = data[0], data[1:]
	p.Data = data
	return
}
