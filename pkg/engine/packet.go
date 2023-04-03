package engine

import (
	"crypto/cipher"
	"errors"
	"fmt"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	PacketMagic     = "NPKT"
	PacketBaseLen   = 4 + 4 + pub.KeyLen + cloak.Len + nonce.IVLen
	PacketHeaderLen = PacketBaseLen + nonce.IDLen + slice.Uint32Len + slice.
		Uint16Len + 1
)

// Packet is the standard format for an encrypted, possibly segmented message
// container with parameters for Reed Solomon Forward Error Correction.
type Packet struct {
	ID      nonce.ID
	To      *pub.Key
	CloakTo cloak.PubKey
	From    *prv.Key
	fromPub *pub.Key
	iv      nonce.IV
	// Seq specifies the segment number of the message, 4 bytes long.
	Seq uint16
	// Parity is the ratio of redundancy with 256.
	Parity byte
	// Length is the length of the full message.
	Length uint32
	// Data is the message.
	Data []byte
}

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

func (x *Packet) Encode(s *Splice) (e error) {
	// log.T.Ln("encoding", reflect.TypeOf(x),
	// x.ID, x.To, x.From, x.Seq, x.Length, x.Parity,
	// )
	iv := nonce.New()
	var start int
	s.Magic4(PacketMagic).
		Check(make(slice.Bytes, 4)).
		Pubkey(pub.Derive(x.From)).
		Cloak(x.To).
		IV(iv).
		StoreCursor(&start).
		Uint16(x.Seq).
		Byte(x.Parity).
		Uint32(x.Length).
		ID(x.ID).
		TrailingBytes(x.Data)
	// log.D.Ln("packet encode", s.SpliceSegments)
	if e = Encipher(s.GetFrom(start), iv, x.From, x.To,
		"packet encode"); fails(e) {
		return
	}
	hash := sha256.Single(s.GetFrom(8))
	s.SetCursor(4).Check(hash[:4])
	return
}

func (x *Packet) Decode(s *Splice) (e error) {
	var magic string
	k := sha256.Single(s.GetFrom(8))
	kk := slice.Bytes(k[:4])
	check := make(slice.Bytes, 4)
	s.
		ReadMagic4(&magic).
		ReadCheck(&check).
		ReadPubkey(&x.fromPub).
		ReadCloak(&x.CloakTo).
		ReadIV(&x.iv)
	kb := string(kk.ToBytes())
	cb := string(check.ToBytes())
	if kb != cb {
		return errors.New("packet integrity check failed")
	}
	if x.fromPub == nil {
		return errors.New("did not parse valid public key")
	}
	return
}

func (x *Packet) Decrypt(prk *prv.Key, s *Splice) (e error) {
	if e = Encipher(s.GetRest(), x.iv, prk, x.fromPub, "packet decode"); fails(e) {
		return
	}
	s.ReadUint16(&x.Seq).
		ReadByte(&x.Parity).
		ReadUint32(&x.Length).
		ReadID(&x.ID)
	// log.D.Ln(s)
	x.Data = s.GetRest()
	return
}

func (x *Packet) Len() int {
	return PacketHeaderLen + len(x.Data)
}

// Packets is a slice of pointers to packets.
type Packets []*Packet

// sort.Interface implementation.

func (p Packets) Len() int           { return len(p) }
func (p Packets) Less(i, j int) bool { return p[i].Seq < p[j].Seq }
func (p Packets) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// EncodePacket creates a Packet, encrypts the payload using the given private from
// key and the public to key, serializes the form, signs the bytes and appends
// the signature to the end.
func EncodePacket(p PacketParams) (pkt []byte, e error) {
	var blk cipher.Block
	if blk = ciph.GetBlock(p.From, p.To, "packet encode"); fails(e) {
		return
	}
	// log.D.Ln("length", p.Length)
	nonc := nonce.New()
	Seq := slice.NewUint16()
	slice.EncodeUint16(Seq, p.Seq)
	Length := slice.NewUint32()
	slice.EncodeUint32(Length, p.Length)
	pkt = make([]byte, slice.SumLen(Seq, Length,
		p.Data)+1+PacketBaseLen+nonce.IDLen)
	// Append pubkey used for encryption key derivation.
	k := pub.Derive(p.From)
	// cloaked := cloak.GetCloak(p.To)
	// Copy nonce, address and key over top of the header.
	var start int
	s := NewSpliceFrom(pkt).
		Magic4(PacketMagic).
		Pubkey(k).
		Cloak(p.To).
		IV(nonc)
	start = s.GetCursor()
	log.D.Ln("start", start)
	s.Uint16(uint16(p.Seq)).
		Byte(byte(p.Parity)).
		Uint32(uint32(p.Length)).
		ID(p.ID)
	copy(s.GetRest(), p.Data)
	log.D.S("prepacket", s.GetAll().ToBytes())
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, s.GetFrom(start))
	return
}

// GetPacketKeys returns the ToHeaderPub field of the message, checks the packet
// checksum and recovers the public key.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be decrypted.
func GetPacketKeys(b []byte) (from *pub.Key, to cloak.PubKey,
	iv nonce.IV, e error) {
	
	pktLen := len(b)
	if pktLen < PacketBaseLen {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			PacketBaseLen, pktLen)
		log.E.Ln(e)
		return
	}
	var prefix string
	NewSpliceFrom(b).
		ReadMagic4(&prefix).
		ReadPubkey(&from).
		ReadCloak(&to).
		ReadIV(&iv)
	if prefix != PacketMagic {
		e = fmt.Errorf("packet magic bytes not found, expected '%v' got'%v'",
			prefix, PacketMagic)
		return
	}
	// log.D.S("position", s.GetCursor())
	return
}

// DecodePacket a packet and return the Packet with encrypted payload and signer's
// public key. This assumes GetPacketKeys succeeded and the matching private key was
// found.
func DecodePacket(b []byte, from *pub.Key, to *prv.Key,
	iv nonce.IV) (p *Packet, e error) {
	
	pktLen := len(b)
	if pktLen < PacketBaseLen {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			PacketBaseLen, pktLen)
		log.E.Ln(e)
		return
	}
	p = &Packet{}
	var blk cipher.Block
	if blk = ciph.GetBlock(to, from, "packet decode"); fails(e) {
		return
	}
	// This decrypts the rest of the packet, which is encrypted for security.
	data := b[PacketBaseLen:]
	ciph.Encipher(blk, iv, data)
	
	s := NewSpliceFrom(b).
		ReadUint16(&p.Seq).
		ReadByte(&p.Parity).
		ReadUint32(&p.Length).
		ReadID(&p.ID)
	p.Data = s.GetRest()
	return
}
