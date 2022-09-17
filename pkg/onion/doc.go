// Package onion contains the functions required to generate onion packets and
// unwrap them one layer at a time.
//
// An onion packet is a specially formed packet in which several layers of data
// are bundled together with layers of encryption between them in order to
// enable sending of information with forwarding instructions and per-hop data
// specific to the task of the onion.
//
// This package provides for three main types of onions:
//
// - Session purchases - acquiring the session seed and cipher for source routed
//   onion hops.
// - Acknowledgment reverse onions - special onions that can be embedded in
//   a layer that can be anonymously returned to the originator of a packet
//   so that the sender can track the progress of the path for latency
//   guarantee, path failure diagnostics, or onion session purchase progress
//   monitoring.
// - Onion packets, which are much larger, 8, 16, 32, 48 or 64kb in size, in
//   which are embedded routing instructions, potentially acknowledgment
//   onions, for onion route circuits and the actual data payload and session
//   bandwidth counters for authentication to sessions.
package onion
