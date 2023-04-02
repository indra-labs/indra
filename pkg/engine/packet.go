package engine

import (
	"crypto/cipher"
	"fmt"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	PacketMagic = "INDR"
)

// Packet is the standard format for an encrypted, possibly segmented message
// container with parameters for Reed Solomon Forward Error Correction.
type Packet struct {
	ID nonce.ID
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq uint16
	// Length is the number of segments in the batch
	Length uint32
	// Parity is the ratio of redundancy. In each 256 segment
	Parity byte
	// Data is the message.
	Data []byte
}

// PacketOverhead is the base overhead on a packet, use GetOverhead to add any extra
// as found in a Packet.
const PacketOverhead = 4 + pub.KeyLen + cloak.Len + nonce.IVLen

// Packets is a slice of pointers to packets.
type Packets []*Packet

// sort.Interface implementation.

func (p Packets) Len() int           { return len(p) }
func (p Packets) Less(i, j int) bool { return p[i].Seq < p[j].Seq }
func (p Packets) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// PacketParams defines the parameters for creating a (split) packet given a set of
// keys, cipher, and data. To, From, Blk and Data are required, Parity is
// optional, set it to define a level of Reed Solomon redundancy on the split
// packets.
type PacketParams struct {
	ID     nonce.ID
	To     *pub.Key
	From   *prv.Key
	Parity int
	Seq    int
	Length int
	Data   []byte
}

// EncodePacket creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func EncodePacket(p PacketParams) (pkt []byte, e error) {
	var blk cipher.Block
	if blk = ciph.GetBlock(p.From, p.To, "packet encode"); fails(e) {
		return
	}
	nonc := nonce.New()
	Seq := slice.NewUint16()
	slice.EncodeUint16(Seq, p.Seq)
	Length := slice.NewUint32()
	slice.EncodeUint32(Length, p.Length)
	pkt = make([]byte, slice.SumLen(Seq, Length,
		p.Data)+1+PacketOverhead+nonce.IDLen)
	// Append pubkey used for encryption key derivation.
	k := pub.Derive(p.From)
	// cloaked := cloak.GetCloak(p.To)
	// Copy nonce, address and key over top of the header.
	s := NewSpliceFrom(pkt).
		Magic4(PacketMagic).
		Pubkey(k).
		Cloak(p.To).
		IV(nonc).
		Byte(byte(p.Parity)).
		Uint16(uint16(p.Seq)).
		Uint32(uint32(p.Length)).
		ID(p.ID)
	copy(pkt[s.GetCursor():], p.Data)
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, pkt[PacketOverhead:])
	return
}

// GetPacketKeys returns the ToHeaderPub field of the message, checks the packet
// checksum and recovers the public key.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be decrypted.
func GetPacketKeys(b []byte) (from *pub.Key, to cloak.PubKey, e error) {
	pktLen := len(b)
	if pktLen < PacketOverhead {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			PacketOverhead, pktLen)
		log.E.Ln(e)
		return
	}
	var k pub.Bytes
	c := new(slice.Cursor)
	prefix := string(b[:c.Inc(4)])
	if prefix != PacketMagic {
		e = fmt.Errorf("packet magic bytes not found, expected '%v' got'%v'",
			prefix, PacketMagic)
		return
	}
	copy(k[:], b[*c:c.Inc(pub.KeyLen)])
	copy(to[:], b[*c:c.Inc(cloak.Len)])
	if from, e = pub.FromBytes(k[:]); fails(e) {
		return
	}
	return
}

// DecodePacket a packet and return the Packet with encrypted payload and signer's
// public key. This assumes GetPacketKeys succeeded and the matching private key was
// found.
func DecodePacket(d []byte, from *pub.Key, to *prv.Key) (p *Packet, e error) {
	pktLen := len(d)
	if pktLen < PacketOverhead {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			PacketOverhead, pktLen)
		log.E.Ln(e)
		return
	}
	p = &Packet{}
	// copy the nonce
	var nonc nonce.IV
	copy(nonc[:], d[4+pub.KeyLen+cloak.Len:PacketOverhead])
	var blk cipher.Block
	if blk = ciph.GetBlock(to, from, "packet decode"); fails(e) {
		return
	}
	// This decrypts the rest of the packet, which is encrypted for security.
	data := d[PacketOverhead:]
	ciph.Encipher(blk, nonc, data)
	
	seq := slice.NewUint16()
	length := slice.NewUint32()
	seq, data = slice.Cut(data, slice.Uint16Len)
	p.Seq = uint16(slice.DecodeUint16(seq))
	length, data = slice.Cut(data, slice.Uint32Len)
	p.Length = uint32(slice.DecodeUint32(length))
	p.Parity, data = data[0], data[1:]
	copy(p.ID[:], data[:nonce.IDLen])
	p.Data = data[nonce.IDLen:]
	return
}
