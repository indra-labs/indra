package dialog

//
// import (
// 	"sync"
//
// 	"github.com/Indra-Labs/indra/pkg/key"
// 	"github.com/Indra-Labs/indra/pkg/old/message"
// )
//
// // Dialog is a data structure for tracking keys used in a message exchange.
// type Dialog struct {
// 	sync.Mutex
// 	// LastIn is the newest pubkey seen in a received message from the
// 	// correspondent.
// 	LastIn *key.Public
// 	// LastOut is the newest privkey used in an outbound message.
// 	LastOut *key.Private
// 	// Seen are the keys that have been seen since the last new message sent
// 	// out to the correspondent.
// 	Seen []*key.Public
// 	// Used are the recently used keys that have not been invalidated by the
// 	// counterparty sending them in the Expires field.
// 	Used []*key.Private
// 	// UsedFingerprints are 1:1 mapped to Used private keys for fast
// 	// recognition. These have been sent in Expires field.
// 	UsedFingerprints []finger.Print
// 	// SegmentSize is the size of packets used in the Dialog. Anything
// 	// larger will be segmented and potentially augmented with Reed Solomon
// 	// parity shards for retransmit avoidance.
// 	SegmentSize uint16
// }
//
// // New creates a new Dialog for tracking a conversation between two nodes.
// // For the initiator, the pubkey is the current one advertised by the
// // correspondent, and for a correspondent, this pubkey is from the first one
// // appearing in the initial message.
// func New(pub *key.Public) (d *Dialog) {
// 	d = &Dialog{LastIn: pub}
// 	return
// }
//
// // Frame is the data format that goes on the wire. This message is wrapped
// // inside a Data and the payload is also inside a Data.
// type Frame struct {
// 	// Expires are the fingerprints of public keys that the correspondent
// 	// can now discard as they will not be used again.
// 	Expires []finger.Print
// 	// Seen are all the keys excluding the To key to signal these can be
// 	// deleted.
// 	Seen []finger.Print
// 	// Data is a Crypt containing a Data.
// 	Data *message.Payload
// }
//
// // Send issues a new message.
// func (d *Dialog) Send(payload []byte) (wf *Frame, e error) {
// 	// generate the sender private key
// 	var prv *key.Private
// 	if prv, e = key.GeneratePrivate(); log.I.Chk(e) {
// 		return
// 	}
// 	// pub := prv.Public()
// 	wf = &Frame{}
// 	// Fill in the 'From' key to the pubkey of the new private key.
// 	// wf.From = pub.ToBytes()
// 	// Lock the mutex of Dialog, so we can update the used/seen keys.
// 	d.Mutex.Lock()
// 	// We always send new messages to the last known correspondent pubkey.
// 	// lastin := d.LastIn
// 	// Move the last outbound private key into the Used field.
// 	if d.LastOut != nil {
// 		d.Used = append(d.Used, d.LastOut)
// 	}
// 	// Set current key to the last used.
// 	d.LastOut = prv
// 	// Collect the used keys to put in the expired. These will be deleted
// 	// in the receiver function.
// 	if len(d.Used) > 0 {
// 		for i := range d.Used {
// 			fp := d.Used[i].Public().Fingerprint()
// 			wf.Expires = append(wf.Expires, fp)
// 		}
// 	}
// 	// Seen keys signal to the correspondent they can discard the related
// 	// private key as it will not be addressed to again.
// 	if len(d.Seen) > 0 {
// 		for i := range d.Seen {
// 			wf.Seen = append(wf.Seen, d.Seen[i].Fingerprint())
// 		}
// 	}
// 	// This is the last access on the Dialog, so we can unlock here.
// 	d.Mutex.Unlock()
// 	// Getting secret and To here outside the critical section as it
// 	// doesn't need locking once the pubkey is copied.
// 	// tofp := lastin.Fingerprint()
// 	// wf.To = &tofp
// 	var msg *message.Payload
// 	if msg, e = message.New(payload, prv); log.E.Chk(e) {
// 		return
// 	}
// 	wf.Data = msg
// 	return
// }
//
// // Receive processes a received message, handles expiring correspondent and
// // prior send keys, and returns the decrypted message to the caller.
// func (d *Dialog) Receive(message []byte) (m *message.Payload, e error) {
// 	// Lock the mutex of Dialog, so we can update the used/seen keys.
// 	d.Mutex.Lock()
// 	// This is the last access on the Dialog, so we can unlock here.
// 	d.Mutex.Unlock()
// 	return
// }
