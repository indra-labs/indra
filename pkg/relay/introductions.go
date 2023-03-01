package relay

import (
	"sync"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/messages/intro"
	"git-indra.lan/indra-labs/indra/pkg/messages/introquery"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Intros map[pub.Bytes]slice.Bytes

type KnownIntros map[pub.Bytes]*intro.Layer

// Introductions is a map of existing known hidden service keys and the
// routing header for requesting a new one on behalf of the client.
//
// After a header is retrieved, the relay sends back a request to the hidden
// service using the headers in this store with the provided public key from the
// client which is then used to encrypt the provided header and prevent the
// introducing relay from also using the provided header.
type Introductions struct {
	sync.Mutex
	Intros
	KnownIntros
}

func NewIntroductions() *Introductions {
	return &Introductions{Intros: make(Intros),
		KnownIntros: make(KnownIntros)}
}

func (in *Introductions) Find(key pub.Bytes) (header slice.Bytes) {
	in.Lock()
	var ok bool
	if header, ok = in.Intros[key]; ok {
	}
	in.Unlock()
	return
}

func (in *Introductions) Delete(key pub.Bytes) (header slice.Bytes) {
	in.Lock()
	var ok bool
	if header, ok = in.Intros[key]; ok {
		delete(in.Intros, key)
	}
	in.Unlock()
	return
}

func (in *Introductions) AddIntro(pk *pub.Key, header slice.Bytes) {
	in.Lock()
	var ok bool
	key := pk.ToBytes()
	if _, ok = in.Intros[key]; ok {
		log.D.Ln("entry already exists for key %x", key)
	} else {
		in.Intros[key] = header
	}
	in.Unlock()
}

func (eng *Engine) SendIntro(id nonce.ID, target *Session, intr *intro.Layer) {
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := HiddenService(id, intr, se[len(se)-1], c, eng.KeySet)
	log.D.Ln("sending out intro onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, func(id nonce.ID, b slice.Bytes) {
		log.I.Ln("received routing header request for %s", intr.Key.ToBase32())
	})
}

func (eng *Engine) gossipIntro(intr *intro.Layer) {
	log.D.F("propagating hidden service introduction for %x", intr.Key.ToBytes())
	done := qu.T()
	msg := make(slice.Bytes, intro.Len)
	c := slice.NewCursor()
	intr.Encode(msg, c)
	nPeers := eng.NodesLen()
	peerIndices := make([]int, nPeers)
	for i := 1; i < nPeers; i++ {
		peerIndices[i] = i
	}
	cryptorand.Shuffle(nPeers, func(i, j int) {
		peerIndices[i], peerIndices[j] = peerIndices[j], peerIndices[i]
	})
	// We broadcast the received introduction to two other randomly selected
	// nodes, which guarantees the entire network will see the intro at least
	// once.
	var cursor int
	for {
		select {
		case <-eng.C.Wait():
			return
		case <-done:
			return
		default:
		}
		n := eng.FindNodeByIndex(peerIndices[cursor])
		n.Transport.Send(msg)
		cursor++
		if cursor > len(peerIndices)-1 {
			break
		}
	}
	log.T.Ln("finished broadcasting intro")
}

func (eng *Engine) intro(intr *intro.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	eng.Introductions.Lock()
	if intr.Validate() {
		if _, ok := eng.Introductions.KnownIntros[intr.Key.ToBytes()]; ok {
			eng.Introductions.Unlock()
			return
		}
		log.D.F("%s storing intro for %s", eng.GetLocalNodeAddress().String(),
			intr.Key.ToBase32())
		eng.Introductions.KnownIntros[intr.Key.ToBytes()] = intr
		log.D.F("%s sending out intro to %s at %s to all known peers",
			eng.GetLocalNodeAddress(), intr.Key.ToBase32(),
			intr.AddrPort.String())
		sender := eng.SessionManager.FindNodeByAddrPort(intr.AddrPort)
		nodes := make(map[nonce.ID]*Node)
		eng.SessionManager.ForEachNode(func(n *Node) bool {
			if n.ID != sender.ID {
				nodes[n.ID] = n
			}
			return false
		})
		counter := 0
		for i := range nodes {
			log.T.F("sending intro to %s", nodes[i].AddrPort.String())
			nodes[i].Transport.Send(b)
			counter++
			if counter < 2 {
				continue
			}
			break
		}
		eng.Introductions.Unlock()
	}
}

func IntroQuery(hsk *pub.Key, client *Session, s Circuit,
	ks *signer.KeySet) Skins {
	
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	n := GenNonces(6)
	var returnNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = s[3].PayloadPub
	pubs[1] = s[4].PayloadPub
	pubs[2] = client.PayloadPub
	return Skins{}.
		ReverseCrypt(s[0], ks.Next(), n[0], 3).
		ReverseCrypt(s[1], ks.Next(), n[1], 2).
		ReverseCrypt(s[2], ks.Next(), n[2], 1).
		IntroQuery(hsk, prvs, pubs, returnNonces).
		ReverseCrypt(s[3], prvs[0], n[3], 3).
		ReverseCrypt(s[4], prvs[1], n[4], 2).
		ReverseCrypt(client, prvs[2], n[5], 1)
}

func (eng *Engine) introquery(iq *introquery.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	eng.Introductions.Lock()
	var ok bool
	var il *intro.Layer
	if il, ok = eng.Introductions.KnownIntros[iq.Key.ToBytes()]; !ok {
		// if the reply is zeroes the querant knows it needs to retry at a
		// different relay
		il = &intro.Layer{}
	}
	eng.Introductions.Unlock()
	header := b[*c:c.Inc(crypt.ReverseHeaderLen)]
	rb := FormatReply(header,
		Encode(il), iq.Ciphers, iq.Nonces)
	switch on1 := prev.(type) {
	case *crypt.Layer:
		sess := eng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.RelayRate * len(b) / 2
			out := sess.RelayRate * len(rb) / 2
			eng.DecSession(sess.ID, in+out, false, "introquery")
		}
	}
	eng.handleMessage(rb, iq)
}
