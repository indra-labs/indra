package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/hiddenservice"
	"git-indra.lan/indra-labs/indra/pkg/messages/intro"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Referrers map[pub.Bytes][]pub.Bytes

func HiddenService(id nonce.ID, il *intro.Layer, client *Session, s Circuit,
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
		HiddenService(id, il, prvs, pubs, returnNonces).
		ReverseCrypt(s[3], prvs[0], n[3], 3).
		ReverseCrypt(s[4], prvs[1], n[4], 2).
		ReverseCrypt(client, prvs[2], n[5], 1)
}

func (eng *Engine) hiddenservice(hs *hiddenservice.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	log.D.F("%s adding introduction for key %s", eng.GetLocalNodeAddress(),
		hs.Layer.Key.ToBase32())
	eng.Introductions.AddIntro(hs.Layer.Key, b[*c:])
	log.I.Ln("stored new introduction, starting broadcast")
	go eng.gossipIntro(&hs.Layer)
}
