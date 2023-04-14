package engine

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"sort"
	"time"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/b32/based32"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	magic2 "git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const AddrLen = net.IPv6len + 2

type NameOffset struct {
	Offset int
	Name   string
	slice.Bytes
}

type SpliceSegments []NameOffset

func (s SpliceSegments) String() (o string) {
	for i := range s {
		o += fmt.Sprintf("%s %d ", s[i].Name, s[i].Offset)
	}
	return
}

func (s SpliceSegments) Len() int           { return len(s) }
func (s SpliceSegments) Less(i, j int) bool { return s[i].Offset < s[j].Offset }
func (s SpliceSegments) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type Splice struct {
	b slice.Bytes
	c *slice.Cursor
	SpliceSegments
}

func (s *Splice) String() (o string) {
	o = "splice:"
	seg := s.GetSlicesFromSegments()
	var prevString string
	for i := range seg {
		switch v := seg[i].(type) {
		case string:
			o += "\n" + v + " "
			prevString = v
		case slice.Bytes:
			if len(v) > 72 {
				o += "\n "
			}
			var oe string
			for j := range v {
				if (j)%4 == 0 && j != 0 {
					oe += ""
				}
				if j == 0 {
					oe += ""
				}
				if v[j] >= '0' && v[j] <= '9' ||
					v[j] >= 'a' && v[j] <= 'z' ||
					v[j] >= 'A' && v[j] <= 'Z' {
					oe += string(v[j])
				} else {
					oe += "."
				}
			}
			if prevString == "magic" {
				o += color.Red.Sprint(oe) + " "
			} else {
				o += color.Gray.Sprint(oe) + " "
				
			}
			if prevString != "remainder" {
				hexed := hex.EncodeToString(v.ToBytes())
				var oHexed string
				var revHex string
				for {
					if len(hexed) >= 8 {
						revHex, hexed = hexed[:8], hexed[8:]
						oHexed += revHex + " "
					} else {
						oHexed += hexed
						break
					}
				}
				o += color.Gray.Sprint(color.Bold.Sprint(oHexed))
			}
			if prevString == "pubkey" {
				var oo string
				var e error
				if oo, e = based32.Codec.Encode(v.ToBytes()); fails(e) {
					o += "<error: " + e.Error() + " >"
				}
				oo = oo[3:]
				tmp := make(slice.Bytes, 0, len(oo))
				tmp = append(tmp[:13], append([]byte("..."),
					tmp[len(tmp)-8:]...)...)
				oo = string(tmp)
				o += color.LightGreen.Sprint(" ", oo)
			}
			if prevString == "ID" {
				var oo string
				var e error
				if oo, e = based32.Codec.Encode(v.ToBytes()); fails(e) {
					o += "<error: " + e.Error() + " >"
				}
				o += color.LightBlue.Sprint(oo[:13])
			}
		}
	}
	return
}

func NewSplice(length int) (splicer *Splice) {
	splicer = &Splice{make(slice.Bytes, length), slice.NewCursor(), SpliceSegments{}}
	return
}

func LoadSplice(b slice.Bytes, c *slice.Cursor) (splicer *Splice) {
	return &Splice{b, c, SpliceSegments{}}
}

func NewSpliceFrom(b slice.Bytes) (splicer *Splice) {
	return LoadSplice(b, slice.NewCursor())
}

func (s *Splice) GetSegments() (segments SpliceSegments) {
	m := make(map[int]NameOffset)
	for i := range s.SpliceSegments {
		m[s.SpliceSegments[i].Offset] = s.SpliceSegments[i]
	}
	s.SpliceSegments = s.SpliceSegments[:0]
	for i := range m {
		s.SpliceSegments = append(s.SpliceSegments, m[i])
	}
	sort.Sort(s.SpliceSegments)
	return s.SpliceSegments
}

func (s *Splice) GetSlicesFromSegments() (segments []interface{}) {
	segs := s.GetSegments()
	var cur, prev int
	for i := range segs {
		cur = segs[i].Offset
		segments = append(segments, segs[i].Name)
		segments = append(segments, s.b[prev:cur])
		prev = cur
	}
	if len(s.b) > cur {
		segments = append(segments, "remainder")
		segments = append(segments, s.b[cur:])
	}
	return
}

func (s *Splice) GetCursor() int { return int(*s.c) }

func (s *Splice) Remaining() int { return s.Len() - s.GetCursor() }

func (s *Splice) Advance(n int, name string) int {
	s.c.Inc(n)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: name})
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

func (s *Splice) GetRoutingHeaderFromCursor() (r RoutingHeaderBytes) {
	rh := s.GetRange(s.GetCursor(), s.Advance(RoutingHeaderLen,
		"routing header"))
	copy(r[:], rh)
	return
}

// GetAll returns the whole of the buffer.
func (s *Splice) GetAll() slice.Bytes { return s.b }

func (s *Splice) GetFrom(p int) slice.Bytes {
	return s.b[p:]
}

func (s *Splice) GetUntil(p int) slice.Bytes {
	return s.b[:p]
}

func (s *Splice) GetRest() slice.Bytes {
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
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "magic"})
	return s
}

func (s *Splice) ReadMagic(out *string) *Splice {
	*out = string(s.b[*s.c:s.c.Inc(magic2.Len)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "magic"})
	return s
}

func (s *Splice) Magic4(magic string) *Splice {
	copy(s.b[*s.c:s.c.Inc(4)], magic)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "magic4"})
	return s
}

func (s *Splice) ReadMagic4(out *string) *Splice {
	*out = string(s.b[*s.c:s.c.Inc(4)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "magic4"})
	return s
}

func (s *Splice) ID(id nonce.ID) *Splice {
	copy(s.b[*s.c:s.c.Inc(nonce.IDLen)], id[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "ID"})
	return s
}

func (s *Splice) ReadID(id *nonce.ID) *Splice {
	copy((*id)[:], s.b[*s.c:s.c.Inc(nonce.IDLen)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "ID"})
	return s
}

func (s *Splice) IV(iv nonce.IV) *Splice {
	copy(s.b[*s.c:s.c.Inc(nonce.IVLen)], iv[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "IV"})
	return s
}

func (s *Splice) Nonces(iv Nonces) *Splice {
	for i := range iv {
		s.IV(iv[i])
	}
	return s
}

func (s *Splice) ReadIV(iv *nonce.IV) *Splice {
	copy((*iv)[:], s.b[*s.c:s.c.Inc(nonce.IVLen)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "IV"})
	return s
}

func (s *Splice) ReadNonces(iv *Nonces) *Splice {
	for i := range iv {
		s.ReadIV(&iv[i])
	}
	return s
}

func (s *Splice) Cloak(pk *crypto.Pub) *Splice {
	if pk == nil {
		s.Advance(crypto.CloakLen, "nil receiver pubkey")
		return s
	}
	to := crypto.GetCloak(pk)
	copy(s.b[*s.c:s.c.Inc(crypto.CloakLen)], to[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "cloak"})
	return s
}

func (s *Splice) ReadCloak(ck *crypto.PubKey) *Splice {
	copy((*ck)[:], s.b[*s.c:s.c.Inc(crypto.CloakLen)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "cloak"})
	return s
}

func (s *Splice) Pubkey(from *crypto.Pub) *Splice {
	if from == nil {
		log.E.Ln("given empty pubkey, doing nothing")
		return s
	}
	pubKey := from.ToBytes()
	copy(s.b[*s.c:s.c.Inc(crypto.PubKeyLen)], pubKey[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "pubkey"})
	return s
}

func (s *Splice) ReadPubkey(from **crypto.Pub) *Splice {
	var f *crypto.Pub
	var e error
	if f, e = crypto.PubFromBytes(s.b[*s.c:s.c.Inc(crypto.PubKeyLen)]); !fails(e) {
		*from = f
	}
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "pubkey"})
	return s
}

func (s *Splice) Prvkey(from *crypto.Prv) *Splice {
	b := from.ToBytes()
	copy(s.b[*s.c:s.c.Inc(crypto.PrvKeyLen)], b[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "prvkey"})
	return s
}

func (s *Splice) ReadPrvkey(out **crypto.Prv) *Splice {
	if f := crypto.PrvKeyFromBytes(s.b[*s.c:s.c.Inc(crypto.PrvKeyLen)]); f == nil {
		return s
	} else {
		*out = f
	}
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "prvkey"})
	return s
}

func (s *Splice) Uint16(v uint16) *Splice {
	slice.EncodeUint16(s.b[*s.c:s.c.Inc(slice.Uint16Len)], int(v))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint16"})
	return s
}

func (s *Splice) ReadUint16(v *uint16) *Splice {
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint16"})
	*v = uint16(slice.DecodeUint16(s.b[*s.c:s.c.Inc(slice.Uint16Len)]))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint16"})
	return s
}

func (s *Splice) Uint32(v uint32) *Splice {
	slice.EncodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)], int(v))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint32"})
	return s
}

func (s *Splice) ReadUint32(v *uint32) *Splice {
	*v = uint32(slice.DecodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)]))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint32"})
	return s
}

func (s *Splice) Uint64(v uint64) *Splice {
	slice.EncodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)], v)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint64"})
	return s
}

func (s *Splice) ReadUint64(v *uint64) *Splice {
	*v = slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "uint64"})
	return s
}

func (s *Splice) ReadMilliSatoshi(v *lnwire.MilliSatoshi) *Splice {
	*v = lnwire.MilliSatoshi(slice.
		DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "mSAT"})
	return s
}

func (s *Splice) Duration(v time.Duration) *Splice {
	slice.EncodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)], uint64(v))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: fmt.Sprint(v)})
	return s
}

func (s *Splice) ReadDuration(v *time.Duration) *Splice {
	*v = time.Duration(slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: fmt.Sprint(*v)})
	return s
}

func (s *Splice) Time(v time.Time) *Splice {
	slice.EncodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)], uint64(v.UnixNano()))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: fmt.Sprint(v)})
	return s
}

func (s *Splice) ReadTime(v *time.Time) *Splice {
	*v = time.Unix(0,
		int64(slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)])))
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: fmt.Sprint(*v)})
	return s
}

func (s *Splice) Hash(h sha256.Hash) *Splice {
	copy(s.b[*s.c:s.c.Inc(sha256.Len)], h[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "hash"})
	return s
}

func (s *Splice) Ciphers(h Ciphers) *Splice {
	for i := range h {
		s.Hash(h[i])
	}
	return s
}

func (s *Splice) ReadHash(h *sha256.Hash) *Splice {
	copy((*h)[:], s.b[*s.c:s.c.Inc(sha256.Len)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "hash"})
	return s
}

func (s *Splice) ReadCiphers(h *Ciphers) *Splice {
	for i := range h {
		s.ReadHash(&h[i])
	}
	return s
}

func (s *Splice) RoutingHeader(b RoutingHeaderBytes) *Splice {
	copy(s.b[*s.c:s.c.Inc(RoutingHeaderLen)], b[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "routingheader"})
	return s
}

func (s *Splice) ReadRoutingHeader(b *RoutingHeaderBytes) *Splice {
	*b = s.GetRoutingHeaderFromCursor()
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "routingheader"})
	return s
}

func (s *Splice) AddrPort(a *netip.AddrPort) *Splice {
	if a == nil {
		s.Advance(AddrLen, "nil AddrPort")
		return s
	}
	var ap []byte
	var e error
	if ap, e = a.MarshalBinary(); fails(e) {
		return s
	}
	s.b[*s.c] = byte(len(ap))
	copy(s.b[s.c.Inc(1):s.c.Inc(AddrLen)], ap)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: color.Yellow.Sprint(a.String())})
	return s
}

func (s *Splice) ReadAddrPort(ap **netip.AddrPort) *Splice {
	*ap = &netip.AddrPort{}
	apLen := s.b[*s.c]
	apBytes := s.b[s.c.Inc(1):s.c.Inc(AddrLen)]
	if e := (*ap).UnmarshalBinary(apBytes[:apLen]); fails(e) {
	}
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c),
			Name: color.Yellow.Sprint((*ap).String())})
	return s
}

func (s *Splice) Byte(b byte) *Splice {
	s.b[*s.c] = byte(b)
	s.c.Inc(1)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "byte"})
	return s
}

func (s *Splice) ReadByte(b *byte) *Splice {
	*b = s.b[*s.c]
	s.c.Inc(1)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "byte"})
	return s
}

func (s *Splice) Bytes(b []byte) *Splice {
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(b))
	copy(s.b[*s.c:s.c.Inc(slice.Uint32Len)], bytesLen)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "32 bit length"})
	copy(s.b[*s.c:s.c.Inc(len(b))], b)
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "bytes"})
	return s
}

func (s *Splice) ReadBytes(b *slice.Bytes) *Splice {
	bytesLen := slice.DecodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "32 bit length"})
	*b = s.b[*s.c:s.c.Inc(bytesLen)]
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "bytes"})
	return s
}

func (s *Splice) Signature(sb *crypto.SigBytes) *Splice {
	copy(s.b[*s.c:s.c.Inc(crypto.SigLen)], sb[:])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "signature"})
	return s
}

func (s *Splice) ReadSignature(sb *crypto.SigBytes) *Splice {
	copy(sb[:], s.b[*s.c:s.c.Inc(crypto.SigLen)])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "signature"})
	return s
}

func (s *Splice) Check(c slice.Bytes) *Splice {
	copy(s.b[*s.c:s.c.Inc(4)], c[:4])
	s.SpliceSegments = append(s.SpliceSegments,
		NameOffset{Offset: int(*s.c), Name: "check"})
	return s
}

func (s *Splice) ReadCheck(c *slice.Bytes) *Splice {
	copy((*c)[:4], s.b[*s.c:s.c.Inc(4)])
	return s
}

func (s *Splice) Done() {}

func (s *Splice) Len() int {
	return len(s.b)
}

func (s *Splice) TrailingBytes(b slice.Bytes) *Splice {
	copy(s.b[*s.c:], b)
	return s
}

func (s *Splice) StoreCursor(c *int) *Splice {
	*c = int(*s.c)
	return s
}
