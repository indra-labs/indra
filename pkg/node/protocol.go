// Package node provides the magic keys that identify each network swarm in the Indra network - mainnet, testnet and simnet.
package node

// Swarm is an Indra network. Encodes a network identifier for mainnet, testnet and simnet.
type Swarm uint32

const (
	// MainNet represents the main indra network.
	MainNet Swarm = 0xd9b4bef9

	// TestNet represents the regression test network.
	TestNet Swarm = 0xdab5bffa

	// SimNet represents the simulation test network.
	SimNet Swarm = 0x12141c16
)
