package sess

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/slice"
)

// Data is a data structure returned from engine.PostAcctOnion that tracks the
// information related to the use of a session.
type Data struct {

	// B is the bytes of data that were sent out.
	B slice.Bytes

	// Sessions are the list of sessions in the circuit.
	Sessions sessions.Sessions

	// Billable is the public keys of the sessions used in the circuit.
	//
	// todo: is this actually used???
	Billable []crypto.PubBytes

	// Ret is the return session, which isn't billable.
	Ret crypto.PubBytes

	// ID is the transmission nonce.ID, as found in PendingResponses.
	ID nonce.ID

	// Port is the well-known port designating the protocol of the message.
	Port uint16

	// PostAcct is a collection of hooks that are to be run on the successful
	// receiving of the response or confirmation and the completion of all relaying
	// from source back to source.
	PostAcct []func()
}
