package splice

import (
	"net"
	"net/netip"
	"time"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const AddrLen = net.IPv6len + 2

type Splicer struct {
	b slice.Bytes
	c *slice.Cursor
}

func Splice(b slice.Bytes, c *slice.Cursor) *Splicer {
	return &Splicer{b, c}
}

func (s *Splicer) Magic(magic slice.Bytes) *Splicer {
	copy(s.b[*s.c:s.c.Inc(magicbytes.Len)], magic)
	return s
}

func (s *Splicer) ReadID(id *nonce.ID) *Splicer {
	copy((*id)[:], s.b[*s.c:s.c.Inc(nonce.IDLen)])
	return s
}

func (s *Splicer) ID(id nonce.ID) *Splicer {
	copy(s.b[*s.c:s.c.Inc(nonce.IDLen)], id[:])
	return s
}

func (s *Splicer) IV(iv nonce.IV) *Splicer {
	copy(s.b[*s.c:s.c.Inc(nonce.IVLen)], iv[:])
	return s
}
func (s *Splicer) ReadIV(iv *nonce.IV) *Splicer {
	copy((*iv)[:], s.b[*s.c:s.c.Inc(nonce.IVLen)])
	return s
}

func (s *Splicer) ReadCloak(ck *cloak.PubKey) *Splicer {
	copy((*ck)[:], s.b[*s.c:s.c.Inc(cloak.Len)])
	return s
}

func (s *Splicer) Cloak(pk *pub.Key) *Splicer {
	to := cloak.GetCloak(pk)
	copy(s.b[*s.c:s.c.Inc(cloak.Len)], to[:])
	return s
}

func (s *Splicer) ReadPubkey(from **pub.Key) *Splicer {
	if f, e := pub.FromBytes(s.b[*s.c:s.c.Inc(pub.KeyLen)]); check(e) {
		return s
	} else {
		*from = f
	}
	return s
}

func (s *Splicer) Pubkey(from *pub.Key) *Splicer {
	pubKey := from.ToBytes()
	copy(s.b[*s.c:s.c.Inc(pub.KeyLen)], pubKey[:])
	return s
}

func (s *Splicer) Prvkey(from *prv.Key) *Splicer {
	b := from.ToBytes()
	copy(s.b[*s.c:s.c.Inc(prv.KeyLen)], b[:])
	return s
}
func (s *Splicer) ReadMilliSatoshi(v *lnwire.MilliSatoshi) *Splicer {
	*v = lnwire.MilliSatoshi(slice.
		DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	return s
}
func (s *Splicer) ReadDuration(v *time.Duration) *Splicer {
	*v = time.Duration(slice.DecodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)]))
	return s
}

func (s *Splicer) Uint64(v uint64) *Splicer {
	slice.EncodeUint64(s.b[*s.c:s.c.Inc(slice.Uint64Len)], v)
	return s
}

func (s *Splicer) Uint32(v uint32) *Splicer {
	slice.EncodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)], int(v))
	return s
}

func (s *Splicer) ReadUint16(v *uint16) *Splicer {
	*v = uint16(slice.DecodeUint16(s.b[*s.c:s.c.Inc(slice.Uint16Len)]))
	return s
}

func (s *Splicer) Uint16(v uint16) *Splicer {
	slice.EncodeUint16(s.b[*s.c:s.c.Inc(slice.Uint16Len)], int(v))
	return s
}

func (s *Splicer) ReadHash(h *sha256.Hash) *Splicer {
	copy((*h)[:], s.b[*s.c:s.c.Inc(sha256.Len)])
	zh := sha256.Hash{}
	copy(s.b[*s.c-sha256.Len:*s.c], zh[:])
	return s
}

func (s *Splicer) Hash(h sha256.Hash) *Splicer {
	copy(s.b[*s.c:s.c.Inc(sha256.Len)], h[:])
	return s
}

func (s *Splicer) AddrPort(a *netip.AddrPort) *Splicer {
	var ap []byte
	var e error
	if ap, e = a.MarshalBinary(); check(e) {
		return s
	}
	s.b[*s.c] = byte(len(ap))
	copy(s.b[s.c.Inc(1):s.c.Inc(AddrLen)], ap)
	return s
}

func (s *Splicer) ReadAddrPort(ap **netip.AddrPort) *Splicer {
	apLen := s.b[*s.c]
	apBytes := s.b[s.c.Inc(1):s.c.Inc(AddrLen)]
	*ap = &netip.AddrPort{}
	if e := (*ap).UnmarshalBinary(apBytes[:apLen]); check(e) {
	}
	return s
}

func (s *Splicer) ReadByte(b *byte) *Splicer {
	*b = s.b[*s.c]
	s.c.Inc(1)
	return s
}

func (s *Splicer) Byte(b byte) *Splicer {
	s.b[*s.c] = byte(b)
	s.c.Inc(1)
	return s
}

func (s *Splicer) ReadBytes(b *slice.Bytes) *Splicer {
	bytesLen := slice.DecodeUint32(s.b[*s.c:s.c.Inc(slice.Uint32Len)])
	*b = s.b[*s.c:s.c.Inc(bytesLen)]
	return s
}

func (s *Splicer) Bytes(b []byte) *Splicer {
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(b))
	copy(s.b[*s.c:s.c.Inc(slice.Uint32Len)], bytesLen)
	copy(s.b[*s.c:s.c.Inc(len(b))], b)
	return s
}

func (s *Splicer) Done() {}
