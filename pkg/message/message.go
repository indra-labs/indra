package message

import (
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/fp"
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

// Form is an in memory structure that is structured more in order to be
// mnemonic for the wire format it represents.
//
// The To, Length and Nonce fields are sized to be 32 bytes total length so the
// Payload is also aligned on 32 byte boundaries.
//
// The signature is a BIP62 format signature which allows public key recovery
// thus avoiding the need to also include the public key for ECDH.
//
// The hash that is required to recover it is the concatenation of the previous
// message segments, 32 bytes from To, Length, Nonce and Payload, in the
// encrypted form, thus also securing the authenticity of the entire packet.
type Form struct {
	// To is the fingerprint of the pubkey used in the ECDH key exchange, 12
	// bytes long.
	To fp.Receiver
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq slice.Length
	// Nonce is the IV for the encryption on the Payload. 16 bytes.
	Nonce nonce.IV
	// Seen is the SHA256 truncated hashes of previous received encryption
	// public keys to indicate they won't be reused and can be discarded.
	Seen []fp.Key
	// Payload is the encrypted message.
	Payload []byte
}

const FormDataMinSize = fp.ReceiverLen + slice.Len + nonce.Size + sig.Len

// Encode creates a Form, encrypts the payload using the given private from key
// and the public to key, serializes the form, signs the bytes and appends the
// signature to the end.
func Encode(to *pub.Key, from *prv.Key, seq int, data []byte,
	seen []fp.Key) (pkt []byte, e error) {

	f := &Form{
		To:    to.ToBytes().Receiver(),
		Nonce: nonce.Get(),
		Seen:  seen,
		Seq:   slice.NewLength(),
	}
	slice.EncodeUint32(f.Seq, seq)
	// This is needed for the decoding.
	SeenCount := []byte{byte(len(seen))}
	// Encrypt the payload
	if e = ciph.Cipher(from, to, f.Nonce, data); check(e) {
		return
	}
	f.Payload = data
	// Append signature to the end of the packet.
	var seenBytes []byte
	for i := range f.Seen {
		seenBytes = append(seenBytes, f.Seen[i][:]...)
	}
	// Assemble the slices into a slice, so we can preallocate capacity and
	// avoid more than one allocation. Signature is copied over the final
	// segment, so empty bytes are allocated for this.
	cat := [][]byte{
		f.To[:],
		f.Seq[:],
		f.Nonce[:],
		SeenCount,
		seenBytes,
		f.Payload,
		sig.New(),
	}
	pktLen := slice.SumLen(cat...)
	pkt = make([]byte, 0, pktLen)
	pkt = slice.Concatenate(cat...)
	// Sign the packet.
	var s sig.Bytes
	if s, e = sig.Sign(from, sha256.Single(pkt)); !check(e) {
		// Signature space is pre-allocated so we copy it.
		copy(pkt[:pktLen-sig.Len], s)
	}
	return
}

// Decode a packet and return the form with encrypted payload and signer's
// public key.
func Decode(pkt []byte) (f *Form, p *pub.Key, e error) {
	pktLen := len(pkt)
	if pktLen < FormDataMinSize {
		// If this isn't checked the slice operations later can
		// hit bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			FormDataMinSize, pktLen)
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
	f = &Form{}
	var Seq, Nonce, SeenCount []byte
	f.To, d = slice.Cut(d, fp.ReceiverLen)
	Seq, d = slice.Cut(d, slice.Len)
	copy(f.Seq[:], Seq)
	Nonce, d = slice.Cut(d, nonce.Size)
	copy(f.Nonce[:], Nonce)
	SeenCount, d = slice.Cut(d, slice.Len)
	sc := slice.DecodeUint32(SeenCount)
	if len(d) < sc*slice.Len {
		e = fmt.Errorf("truncated packet")
		log.E.Ln(e)
		return

	}
	var sn []byte
	f.Seen = make([]fp.Key, sc)
	for i := 0; i < sc; i++ {
		sn, d = slice.Cut(d, fp.Len)
		copy(f.Seen[i][:], sn)
	}
	if len(d) == 0 {
		e = fmt.Errorf("empty message payload")
		log.E.Ln(e)
		return
	}
	f.Payload = d
	return
}

// Decrypt the Payload in a Form using a given key pair using ECDH to derive the
// cipher. The receiver must have the private key matched to the To field, the
// public key required is embedded in the signature.
//
// Note: calling this twice will return the packet to its encrypted form.
func (f *Form) Decrypt(to *prv.Key, from *pub.Key) (e error) {
	return ciph.Cipher(to, from, f.Nonce, f.Payload)
}
