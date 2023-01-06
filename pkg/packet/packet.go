// Package packet provides a standard message binary serialised data format and
// message segmentation scheme which includes address.Sender cloaked public
// key and address.Receiver private keys for generating a shared cipher and applying
// to messages/message segments.
package packet

import (
	"crypto/cipher"
	"fmt"
	"time"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ciph"
	"github.com/indra-labs/indra/pkg/key/address"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
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
	// Deadline is a time after which the message should be received and
	// dispatched.
	Deadline time.Time
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
const Overhead = 4 + nonce.IVLen + pub.KeyLen

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
//
// This library is for creating segmented, FEC redundancy protected network
// packets, and the To sender key should be the publicly advertised public key
// of a relay.
//
// Deadline is a special field that gives a timeout period after which an
// incomplete message can be considered expired and flushed from the cache. It
// is 32 bits in size as precision to the second is sufficient, and low latency
// messages will potentially beat the deadline at one second.
type EP struct {
	To       *address.Sender
	From     *prv.Key
	Parity   int
	Seq      int
	Length   int
	Deadline time.Time
	Data     []byte
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
	if blk = ciph.GetBlock(ep.From, ep.To.Key); check(e) {
		return
	}
	nonc := nonce.New()
	Seq := slice.NewUint16()
	slice.EncodeUint16(Seq, ep.Seq)
	Length := slice.NewUint32()
	slice.EncodeUint32(Length, ep.Length)
	Deadline := slice.NewUint64()
	slice.EncodeUint64(Deadline, uint64(ep.Deadline.Unix()))
	pkt = make([]byte, slice.SumLen(Seq, Length, Deadline, ep.Data)+1+Overhead)
	// Append pubkey used for encryption key derivation.
	k := pub.Derive(ep.From).ToBytes()
	// Copy nonce, address and key over top of the header.
	c := new(slice.Cursor)
	copy(pkt[c.Inc(4):c.Inc(nonce.IVLen)], nonc[:])
	copy(pkt[*c:c.Inc(pub.KeyLen)], k[:])
	copy(pkt[*c:c.Inc(slice.Uint16Len)], Seq)
	copy(pkt[*c:c.Inc(slice.Uint32Len)], Length)
	copy(pkt[*c:c.Inc(slice.Uint64Len)], Deadline)
	pkt[*c] = byte(ep.Parity)
	copy(pkt[c.Inc(1):], ep.Data)
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, pkt[Overhead:])
	// last but not least, the packet check header, which protects the
	// entire packet.
	checkBytes := sha256.Single(pkt[4:])
	copy(pkt[:4], checkBytes[:4])
	return
}

// GetKeys returns the To field of the message in order, checks the packet
// checksum and recovers the public key.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be decrypted, and the Decode function will then
// decode a OnionSkin.
func GetKeys(d []byte) (from *pub.Key, e error) {
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
	var k pub.Bytes
	var chek []byte
	c := new(slice.Cursor)
	chek = d[:c.Inc(4)]
	copy(k[:], d[c.Inc(nonce.IVLen):c.Inc(pub.KeyLen)])
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
	f = &Packet{}
	// copy the nonce
	var nonc nonce.IV
	c := new(slice.Cursor)
	copy(nonc[:], d[c.Inc(4):c.Inc(nonce.IVLen)])
	var blk cipher.Block
	if blk = ciph.GetBlock(to, from); check(e) {
		return
	}
	// This decrypts the rest of the packet, which is encrypted for
	// security.
	data := d[c.Inc(pub.KeyLen):]
	ciph.Encipher(blk, nonc, data)
	seq := slice.NewUint16()
	length := slice.NewUint32()
	deadline := slice.NewUint32()
	seq, data = slice.Cut(data, slice.Uint16Len)
	f.Seq = uint16(slice.DecodeUint16(seq))
	length, data = slice.Cut(data, slice.Uint32Len)
	f.Length = uint32(slice.DecodeUint32(length))
	deadline, data = slice.Cut(data, slice.Uint64Len)
	f.Deadline = time.Unix(int64(slice.DecodeUint64(deadline)), 0)
	f.Parity, data = data[0], data[1:]
	f.Data = data
	return
}
