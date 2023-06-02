package cfg

// SeedAddress is a form of the key and network address for seeds on a network.
type SeedAddress struct {

	// ID is the p2p identifier (peer identity key).
	ID string

	// DNSAddress is the hostname of the seed node
	DNSAddress string
}

func NewSeedAddress(dns string, id string) *SeedAddress {

	return &SeedAddress{
		ID:         id,
		DNSAddress: dns,
	}
}
