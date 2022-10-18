package sifr

import (
	"sync"

	"github.com/Indra-Labs/indra/pkg/schnorr"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

// Dialog is a data structure for tracking keys used in a message exchange.
type Dialog struct {
	sync.Mutex
	// LastIn is the newest pubkey seen in a received message from the
	// correspondent.
	LastIn *schnorr.Pubkey
	// LastOut is the newest privkey used in an outbound message.
	LastOut *schnorr.Privkey
	// Seen are the keys that have been seen since the last new message sent
	// out to the correspondent.
	Seen []*schnorr.Pubkey
	// Used are the recently used keys that have not been invalidated by the
	// counterparty sending them in the Expires field.
	Used []*schnorr.Privkey
	// UsedFingerprints are 1:1 mapped to Used private keys for fast
	// recognition. These have been sent in Expires field.
	UsedFingerprints []schnorr.Fingerprint
}

// NewDialog creates a new Dialog for tracking a conversation between two nodes.
// For the initiator, the pubkey is the current one advertised by the
// correspondent, and for a correspondent, this pubkey is from the first one
// appearing in the initial message.
func NewDialog(pub *schnorr.Pubkey) (d *Dialog) {
	d = &Dialog{LastIn: pub}
	return
}

// Frame is the data format that goes on the wire. This message is wrapped
// inside a Message and the payload is also inside a Message.
type Frame struct {
	// To is the fingerprint of the pubkey used in the ECDH key exchange.
	To *schnorr.Fingerprint
	// From is the pubkey corresponding to the private key used in the ECDH
	// key exchange.
	From *schnorr.PubkeyBytes
	// Expires are the fingerprints of public keys that the correspondent
	// can now discard as they will not be used again.
	Expires []schnorr.Fingerprint
	// Seen are all the keys excluding the To key to signal these can be
	// deleted.
	Seen []schnorr.Fingerprint
	// Data is the payload of the message, which is wrapped in a
	// Message.
	Data Message
}

func (d *Dialog) Message(payload []byte) (wf *Frame, e error) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	// We always send new messages to the last known correspondent pubkey.
	tofp := d.LastIn.Fingerprint()
	wf = &Frame{To: &tofp}
	var prv *schnorr.Privkey
	// generate the sender private key
	if prv, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
		return
	}
	// Move the last outbound private key into the Used field.
	if d.LastOut != nil {
		d.Used = append(d.Used, d.LastOut)
	}
	// Set current key to the last used.
	d.LastOut = prv
	// Fill in the 'From' key to the pubkey of the new private key.
	pub := prv.Pubkey()
	wf.From = pub.Serialize()
	if len(d.Used) > 0 {
		for i := range d.Used {
			wf.Expires = append(wf.Expires,
				d.Used[i].Pubkey().Fingerprint())
		}
	}
	if len(d.Seen) > 0 {
		for i := range d.Seen {
			wf.Seen = append(wf.Seen, d.Seen[i].Fingerprint())
		}
	}
	secret := prv.ECDH(d.LastIn)
	var msg *Message
	if msg, e = NewMessage(payload, prv); log.E.Chk(e) {
		return
	}
	var em *Crypt
	if em, e = NewCrypt(msg, secret); log.E.Chk(e) {
		return
	}
	wm := &Message{Payload: em.Serialize()}
	hash := sha256.Hash(wm.Payload)
	var sig *schnorr.Signature
	if sig, e = prv.Sign(hash); log.E.Chk(e) {
		return
	}
	wm.Signature = *sig.Serialize()
	wf.Data = *wm
	return
}
