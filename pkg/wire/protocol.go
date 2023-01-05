package wire

type IndraNet uint32

const (
	// MainNet represents the main indra network.
	MainNet IndraNet = 0xd9b4bef9

	// TestNet represents the regression test network.
	TestNet IndraNet = 0xdab5bffa

	// SimNet represents the simulation test network.
	SimNet IndraNet = 0x12141c16
)
