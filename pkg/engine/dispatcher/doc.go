// Package dispatcher is a network packet send/receive handler for peer to peer connections between relays.
//
// Messages between peers are usually somewhat large, multi-layered onion messages that contain forwarding instructions for sending, and the dispatcher breaks them down into uniform sized segments and randomises their order.
//
// On the receiving side, there is a buffer for incoming message segments, and when sufficient segments are received to enable reconstruction, reconstruction is attempted.
//
// Messages are broken up into pieces with additional segments added to ensure the receiver gets enough pieces to decode the message without a message retransmit request, using Reed Solomon encoding (accelerated by AVX).
//
// The dispatcher operates with reliable TCP connections, and does not directly influence retransmit but instead monitors the latency of the messages and identifies when there has been a retransmit and increases the redundant data added to the stream of message packet segments.
//
// In this way, the dispatcher aims to always see sufficient data arrive in one message cycle so to minimise the latency of connections between peers, and the subsequent latency of client's routed packets.
//
// Another feature of the dispatcher is key change processing - this is implemented as a concurrent update from the receiver specifying a new public key to use, it keeps the old key for a time to deal with in transit messages that were encrypted with the old key and the peer should update its key to use for all messages after the key has been received and updated.
package dispatcher
