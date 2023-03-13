package octet

import (
	"net"
	"net/netip"
	"time"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/sig"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	magic2 "git-indra.lan/indra-labs/indra/pkg/engine/magic"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const AddrLen = net.IPv6len + 2

type Splice struct {
	b slice.Bytes
	c *slice.Cursor
}

func New(length int) (splicer *Splice) {
	splicer = &Splice{make(slice.Bytes, length), slice.NewCursor()}
	return
}

func Load(b slice.Bytes, c *slice.Cursor) (splicer *Splice) {
	return &Splice{b, c}
}

func (s *Splice) GetCursor() int { return int(*s.c) }

func (s *Splice) Remaining() int { return s.Len() - s.GetCursor() }

func (s *Splice) Advance(n int) int {
	s.c.Inc(n)
	return int(*s.c)
}

func (s *Splice) Rewind(n int) int {
	newLoc := int(*s.c) - n
	if newLoc < 0 {
		newLoc = 0
	}
	*s.c = slice.Cursor(newLoc)
	return newLoc
}

func (s *Splice) SetCursor(c int) *Splice {
	*s.c = slice.Cursor(c)
	return s
}

// GetRange slices out a segment of the splicer's data bytes and returns it.
// Use -1 to indicate start or beginning respectively, or both to get all of it.
func (s *Splice) GetRange(start, end int) slice.Bytes {
	switch {
	case start == -1 && end == -1:
		return s.b
	case start == -1:
		return s.b[:end]
	case end == -1:
		return s.b[start:]
	}
	return s.b[start:end]
}

func (s *Splice) GetCursorToEnd() slice.Bytes {
	return s.b[s.GetCursor():]
}

func (s *Splice) CopyRanges(start1, end1, start2, end2 int) {
	copy(s.GetRange(start1, end1), s.GetRange(start2, end2))
}

func (s *Splice) CopyIntoRange(b slice.Bytes, start, end int) {
	copy(s.GetRange(start, end), b[:end-start])
}

func (s *Splice) Magic(magic string) *Splice {
	copy(s.b[*s.c:s.c.Inc(magic2.Len)], magic)
	return s
}

func (s *Splice) ReadMagic(out *string) *Splice {
	*out = string(s.b[*s.c:s.c.Inc(magic2.Len)])
	return s
}

func (s *Splice) ID(id nonce.ID) *Splice {
	copy(s.b[*s.c:s.c.Inc(nonce.IDLen)], id[:])
	return s
}

func (s *Splice) ReadID(id *nonce.ID) *Splice {
	copy((*id)[:], s.b[*s.c:s.c.Inc(nonce.IDLen)])
	return s
}

func (s *Splice) IV(iv nonce.IV) *Splice {
	copy(s.b[*s.c:s.c.Inc(nonce.IVLen)], iv[:])
	return s
}

func (s *Splice) IVTriple(h [3]nonce.IV) *Splice {
	for i := range h {
		s.IV(h[i])
	}
	return s
}

func (s *Splice) ReadIV(iv *nonce.IV) *Splice {
	copy((*iv)[:], s.b[*s.c:s.c.Inc(nonce.IVLen)])
	return s
}

func (s *Splice) ReadIVTriple(h *[3]nonce.IV) *Splice {
	for i := range h {
		s.ReadIV(&h[i])
	}
	return s
}

func (s *Splice) Cloak(pk *pub.Key) *Splice {
	to := cloak.GetCloak(pk)
	copy(s.b[*s.c:s.c.Inc(cloak.Len)], to[:])
	return s
}

func (s *Splice) ReadCloak(ck *cloak.PubKey) *Splice {
	copy((*ck)[:], s.b[*s.c:s.c.Inc(cloak.Len)])
	return s
}

func (s *Splice) Pubkey(from *pub.Key) *Splice {
	pubKey := from.ToBytes()
	copy(s.b[*s.c:s.c.Inc(pub.KeyLen)], pubKey[:])
	return s
}

func (s *Splice) ReadPubkey(from **pub.Key) *Splice {
	var f *pub.Key
	var e error
	if f, e = pub.FromBytes(s.b[*s.c:s.c.Inc(pub.KeyLen)]); !check(e) {
		*from = f
	}
	return s
}

func (s *Splice) Prvkey(from *prv.Key) *Splice {
	b := from.ToBytes()
	copy(s.b[*s.c:s.c.Inc(prv.KeyLen)], b[:])
	return s
}

func (s *Splice) ReadPrvkey(out **prv.Key) *Splice {
	if f := prv.PrivkeyFromBytes(s.b[*s.c:s.c.Inc(prv.KeyLen)]); f == nil {
		return s
	} else {
		*out = f
	}
	return s
}

func (s *Splice) Uint16(v uint16) *Splice {
	slice.EncodeUint16(s.b[*s.c:s.c.Inc(slice.Uint16Len)], int(v))
	return s
}

func (s *Splice) ReadUint16(v *uint16) *Splice {
	*v = uint16(slice.DecodeUint16(s.b[*s.c:s.c.Inc(slice.Uint16Len)]))
	return s
}

func (s *Splice) Uint32(v uint32) *Splice {
	slice.EncodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)], int(v))
	return s
}

func (s *Splice) ReadUint32(v *uint16) *Splice {
	*v = uint16(slice.DecodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)]))
	return s
}

func (s *Splice) Uint64(v uint64) *Splice {
	slice.EncodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)], v)
	return s
}

func (s *Splice) ReadUint64(v *uint16) *Splice {
	*v = uint16(slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	return s
}

func (s *Splice) ReadMilliSatoshi(v *lnwire.MilliSatoshi) *Splice {
	*v = lnwire.MilliSatoshi(slice.
		DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	return s
}

func (s *Splice) ReadDuration(v *time.Duration) *Splice {
	*v = time.Duration(slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	return s
}

func (s *Splice) ReadTime(v *time.Time) *Splice {
	*v = time.Unix(0,
		int64(slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)])))
	return s
}

func (s *Splice) Hash(h sha256.Hash) *Splice {
	copy(s.b[*s.c:s.c.Inc(sha256.Len)], h[:])
	return s
}

func (s *Splice) HashTriple(h [3]sha256.Hash) *Splice {
	for i := range h {
		s.Hash(h[i])
	}
	return s
}

func (s *Splice) ReadHash(h *sha256.Hash) *Splice {
	copy((*h)[:], s.b[*s.c:s.c.Inc(sha256.Len)])
	zh := sha256.Hash{}
	copy(s.b[*s.c-sha256.Len:*s.c], zh[:])
	return s
}

func (s *Splice) ReadHashTriple(h *[3]sha256.Hash) *Splice {
	for i := range h {
		s.ReadHash(&h[i])
	}
	return s
}

func (s *Splice) AddrPort(a *netip.AddrPort) *Splice {
	var ap []byte
	var e error
	if ap, e = a.MarshalBinary(); check(e) {
		return s
	}
	s.b[*s.c] = byte(len(ap))
	copy(s.b[s.c.Inc(1):s.c.Inc(AddrLen)], ap)
	return s
}

func (s *Splice) ReadAddrPort(ap **netip.AddrPort) *Splice {
	apLen := s.b[*s.c]
	apBytes := s.b[s.c.Inc(1):s.c.Inc(AddrLen)]
	*ap = &netip.AddrPort{}
	if e := (*ap).UnmarshalBinary(apBytes[:apLen]); check(e) {
	}
	return s
}

func (s *Splice) Byte(b byte) *Splice {
	s.b[*s.c] = byte(b)
	s.c.Inc(1)
	return s
}

func (s *Splice) ReadByte(b *byte) *Splice {
	*b = s.b[*s.c]
	s.c.Inc(1)
	return s
}

func (s *Splice) Bytes(b []byte) *Splice {
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(b))
	copy(s.b[*s.c:s.c.Inc(slice.Uint32Len)], bytesLen)
	copy(s.b[*s.c:s.c.Inc(len(b))], b)
	return s
}

func (s *Splice) ReadBytes(b *slice.Bytes) *Splice {
	bytesLen := slice.DecodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)])
	*b = s.b[*s.c:s.c.Inc(bytesLen)]
	return s
}

func (s *Splice) Signature(sb *sig.Bytes) *Splice {
	copy(s.b[*s.c:s.c.Inc(sig.Len)], sb[:])
	return s
}

func (s *Splice) ReadSignature(sb *sig.Bytes) *Splice {
	copy(sb[:], s.b[*s.c:s.c.Inc(sig.Len)])
	return s
}

func (s *Splice) Done() {}

func (s *Splice) Len() int {
	return len(s.b)
}
