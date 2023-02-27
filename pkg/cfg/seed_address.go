package cfg

type SeedAddress struct {

	// ID is the p2p identifier
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
