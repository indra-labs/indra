package cfg

type DNSSeedAddress struct {

	// ID is the p2p identifier
	ID string

	// DNSAddress is the hostname of the seed node
	DNSAddress string
}

func NewSeedAddress(dns string, id string) *DNSSeedAddress {

	return &DNSSeedAddress{
		ID:         id,
		DNSAddress: dns,
	}
}
