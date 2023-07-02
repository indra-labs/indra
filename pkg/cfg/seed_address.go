package cfg

// SeedAddress is a form of the key and network address for seeds on a network.
type SeedAddress struct {

	// ID is the p2p identifier (peer identity key).
	ID string

	// DNSAddress is the hostname of the seed node
	DNSAddress string
}

// NewSeedAddress creates a new seed from a network address and seed public key in base58 encoding.
func NewSeedAddress(dns string, id string) *SeedAddress {

	return &SeedAddress{
		ID:         id,
		DNSAddress: dns,
	}
}
