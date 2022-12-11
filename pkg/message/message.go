// Package message provides a standard message binary serialised data format and
// message segmentation scheme which includes address.Sender cloaked public
// key and address.Receiver private keys for generating a shared cipher and applying
// to messages/message segments.
package message

import (
	"crypto/cipher"
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Message is the standard format for an encrypted, possibly segmented message
// container with parameters for Reed Solomon Forward Error Correction and
// contains previously seen cipher keys so the correspondent can free them.
type Message struct {
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq uint16
	// Length is the number of segments in the batch
	Length uint32
	// Parity is the ratio of redundancy. In each 256 segment
	Parity byte
	// Data is the message.
	Data []byte
}

// GetOverhead returns the packet frame overhead given the settings found in the
// packet.
func (p *Message) GetOverhead() int {
	return Overhead
}

// Overhead is the base overhead on a packet, use GetOverhead to add any extra
// as found in a Message.
const Overhead = slice.Uint16Len +
	slice.Uint32Len + 1 + KeyEnd

type Addresses struct {
	To   *address.Sender
	From *prv.Key
}

func Address(To *address.Sender, From *prv.Key) *Addresses {
	return &Addresses{To: To, From: From}
}

const (
	CheckEnd   = 4
	NonceEnd   = CheckEnd + nonce.IVLen
	AddressEnd = NonceEnd + address.Len
	KeyEnd     = AddressEnd + pub.KeyLen
)

// Encode creates a Message, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func Encode(To *address.Sender, From *prv.Key, d []byte) (pkt []byte,
	e error) {

	var blk cipher.Block
	if blk = ciph.GetBlock(From, To.Key); check(e) {
		return
	}
	nonc := nonce.New()
	var to address.Cloaked
	to, e = To.GetCloak()
	Length := slice.NewUint32()
	slice.EncodeUint32(Length, len(d))
	// Concatenate the message pieces together into a single byte slice.
	pkt = slice.Cat(
		// f.Nonce[:],    // 16 bytes \
		// f.To[:],       // 8 bytes   |
		make([]byte, KeyEnd),
		Length, // 4 bytes
		d,
	)
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, pkt[KeyEnd:])
	// Sign the packet.
	var pubKey pub.Bytes
	pubKey = pub.Derive(From).ToBytes()
	// Copy nonce, address, check and signature over top of the header.
	copy(pkt[CheckEnd:NonceEnd], nonc[:])
	copy(pkt[NonceEnd:AddressEnd], to[:])
	copy(pkt[AddressEnd:KeyEnd], pubKey[:])
	// last bot not least, the packet check header, which protects the
	// entire packet.
	checkBytes := sha256.Single(pkt[CheckEnd:])
	copy(pkt[:CheckEnd], checkBytes[:CheckEnd])
	return
}

// GetKeys returns the To field of the message in order, checks the packet
// checksum and recovers the public key signing it.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be decrypted, and the Decode function will then
// decode a Message.
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
	copy(to[:], d[NonceEnd:AddressEnd])
	// split off the signature and recover the public key
	var chek []byte
	chek = d[:CheckEnd]
	checkHash := sha256.Single(d[CheckEnd:])
	if string(chek) != string(checkHash[:4]) {
		e = fmt.Errorf("check failed: got '%v', expected '%v'",
			chek, checkHash[:4])
		return
	}
	if from, e = pub.FromBytes(d[AddressEnd:KeyEnd]); check(e) {
		return
	}
	return
}

// Decode a packet and return the Message with encrypted payload and signer's
// public key. This assumes GetKeys succeeded and the matching private key was
// found.
func Decode(d []byte, from *pub.Key, to *prv.Key) (f *Message, e error) {
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

	f = &Message{}
	// copy the nonce
	var nonc nonce.IV
	copy(nonc[:], d[CheckEnd:NonceEnd])
	var blk cipher.Block
	if blk = ciph.GetBlock(to, from); check(e) {
		return
	}
	// This decrypts the rest of the packet, which is encrypted for
	// security.
	data := d[KeyEnd:]
	ciph.Encipher(blk, nonc, data)
	var length slice.Size32
	length, data = slice.Cut(data, slice.Uint32Len)
	f.Length = uint32(slice.DecodeUint32(length))
	f.Data = data
	// log.I.Ln("decode length", len(data), "length prefix", f.Length)
	return
}
