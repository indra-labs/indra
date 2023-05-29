package rpc

import (
	"github.com/lightningnetwork/lnd/macaroons"
	"gopkg.in/macaroon.v2"
)

type MacaroonCredential struct{ macaroons.MacaroonCredential }

// RequireTransportSecurity implements the PerRPCCredentials interface.
func (m MacaroonCredential) RequireTransportSecurity() bool {
	return false
}

// NewMacaroonCredential returns a copy of the passed macaroon wrapped in a
// MacaroonCredential struct which implements PerRPCCredentials.
func NewMacaroonCredential(m *macaroon.Macaroon) (MacaroonCredential, error) {
	ms := MacaroonCredential{}

	// The macaroon library's Clone() method has a subtle bug that doesn't
	// correctly clone all caveats. We need to use our own, safe clone
	// function instead.
	var err error
	ms.Macaroon, err = macaroons.SafeCopyMacaroon(m)
	if err != nil {
		return ms, err
	}

	return ms, nil
}
