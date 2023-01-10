# Indranet Protocol White Paper

![logo](logo.png)

#### Onion routed distributed virtual private network protocol with anonymised payments to create scaling incentives.
> [David Vennik](mailto:david@cybriq.systems) September 2022 - January 2023

## Abstract

The state of counter-surveillance technologies has remained largely unchanged in
the 20 years since the inception of the [Tor network](https://torproject.org).

The primary use case has always been obscuring the location information of users
from clear net sites, and the more it has been used for this purpose, the more
hostile clear net sites have become towards this network, due to its frequent use
to launch attacks on web services.

With the increasing amounts of value being transported in data packets on the
Internet since the appearance of the Bitcoin network, the need for eliminating
the risks of geographical correlation between payments and user locations
continues to rise.

However, without any way for users to pay routers without creating an audit
trail, the anonymising networks have not grown in nearly a decade, and
thus well heeled attackers have largely been able to keep pace and pluck off
high value targets, such as
the [Carnegie Mellon University](https://blog.torproject.org/did-fbi-pay-university-attack-tor-users/) - implicated in part of what led
to the arrest of the Silk Road founder, Ross Ulbricht.

It is the central thesis of this paper to demonstrate how obfuscating
correlation between payments and session usage can be achieved and create a
marketplace in routing services which can economically increase to a size that
is beyond the capabilities of a state sized actor to fund an attack, while also
improving latency and stability of routed connections.

## Tor Isn't Scaling, But Bitcoin Needs Onion Routing

For comparison, this is Bitcoin's node count:

![image-20220912120917831](image-20220912120917831.png)

Versus Tor in a comparable period:

![tor relay count](torrelaycount.png)

It is not hard to see:

- Tor is not growing, it has flat-lined.

- Bitcoin is growing.

Not only that, you can also see that onion routing is forming an increasingly
large component of Bitcoin connectivity.

## How Indranet Improves Upon Existing Mixnet Designs

Indranet, in contrast to other anonymising network designs, is pure [source routed](https://en.wikipedia.org/wiki/Source_routing). This means that only the clients know the full path along which their traffic moves, and relays have no influence on the path of traffic aside from failing to deliver it. It is similar in many ways to the [HORNET](./1507.05724v1.pdf) mixnet protocol.

Many of the [vulnerabilities](https://en.wikipedia.org/wiki/Mix_network#Vulnerabilities) of mixnets relate to the relays having the ability to change the path of traffic. In Indranet, relays either forward traffic as instructed, or not. To some degree, source routing creates protection against byzantine type faults, because failure to deliver and most malicious attack methods result in clients distrusting the malicious nodes, the same way as they distrust unreliable nodes. Because unreliable nodes cost clients money, effectively, malicious nodes will be quickly forgotten and end up being used by nobody.

### Active Attacks 

It is not possible for adversaries to modify packets without effectively breaking the path, as messages are integrity protected by cryptography that is opaque to other than the client and the at each layer, the relay that will forward or process the message. This can adversely effect the randomly selected next path because it may be inferred that possibly the receiver is faulty rather than the sender. The client will use probes to check the liveness and latency of the path, and by a process of elimination it will determine the sender is faulty and deprecate it from further use, potentially eventually fully blacklisting it and not sharing the node's existence with peers.

### Artificial Gap (packet dropping/delay) and Artificial Bursts

These attacks fail because the client will determine via ping probes that nodes are failing to execute the forwarding/returning of packets and this will gradually result in the malicious nodes not being shared by clients altogether, not just way down the scale of reliability. It is part of the contract of operating a relay on Indranet that your node follows the instructions given to it.

### Timing Analysis

Discovering the relationships between clients and the services they are connecting to is the essence of all of the attacks on mixnets. Indra takes several approaches to resolving this issue:

- **Packet scheduling is shuffled constantly** - Relays do not simply behave as First In First Out buffers, but rather, mix together the messages they receive. All messages have to be segmented for network transport, and as these segments are fed into the outbound queue, they are shuffled so that the ordering is broken up. This slightly increases latency but it decreases associativity between streams of packets constituting messages.
- **Deliberate delays** - This is a tactic that is more often seen in email mixnets but can be used to a degree in lower latency mixnets like Indranet because not all traffic is time critical. Bitcoin transactions and blocks are relatively time critical, but DNS requests can comfortably take up to 100 ms without disrupting the functioning of ancillary centralised type services using them (IE, DNS for web traffic). Email is also a potential service type on Indranet, and because it is inherently slow and non-interactive, the delays can be even longer, potentially several seconds to deliver, and so the segments of the packets as they pass through will not have strong timing correlation. The delays are defined by the client, and raise the cost of the data from a regular non-delayed message by a percentage in proportion with the length of the delay.
- **Client-side anonymity** - Indranet does not also expose relays to uncompensated exposure to entities outside of the network. Potentially Indra can implement Tor style web browser anonymisation, but because such traffic is often used to attack websites and web services it is usually going to be only a matter of time if a free cloak for malicious traffic becomes blacklisted by inbound connection filters on these services. Most Indranet traffic will be passed to decentralised servers, such as Bitcoin, Lightning and IPFS, which do not have a central point of observation either.

## Why We Need Indranet

Three key elements of the Tor protocol make it less than desirable in general.

1. **High Latency** - Establishment of circuits is quite slow, taking a large number of steps
   to "telescope" into a circuit.

2. **Low Reliability** - Once a circuit is running, when it fails, the failure is opaque to the client
   side, and there is no way to provide a latency guarantee or connection
   stability. It is unsuitable for interactive *and* long living connections.

3. **Poor Scaling Incentives** - There is no profit motive to drive expansion of relaying
   capacity, weakening anonymity by not expanding the anonymity set to cope with a rise in the number of users. As the charts showed in the foregoing, there is around 8,000 nodes, of which 6,000 are relays and the remainder private bridges, but the average daily user count of Tor is around 100,000 users. Both numbers could be a lot higher if running a relay wasn't a money losing exercise, and if the system could handle interactive grade latency and very long living connections.

Tor is a poor solution for a very limited subset of the use cases that
benefit from the security of route obfuscation. Indra aims to provide what Tor
has definitely now failed to achieve for a large majority of internet users:
location privacy.

Indranet does not aim to compete with Tor for the use case of tunneling out to clear-net websites and services: the focus is on obscuring the source of traffic within decentralised, peer to peer protocols like Bitcoin, Lightning Network, Bittorrent, IPFS, and other similar, decentralised protocols.

## General Principles of Indranet Protocol

There is four main types of traffic in Indranet:

1. **Network Protocol Chatter** - sharing lists of known network nodes, their advertised exit services, and collaboratively generated statistics on bandwidth and uptime, and their long lived public keys for session initiation.

2. **Purchase of bandwidth sessions** - Combining the use of Lightning Network to perform payments to proxy nodes, and specially formed layered encryption of messages, enabling clients to acquire tokens that grant users the ability to relay arbitrary traffic through relays.

3. **Liveness diagnostics** - In order to keep track of the state of relays on the network, clients send out regular 3 hop messages that circle back in a diamond-shaped topology that they use to acquire the liveness state of relays. This requires the relays to actually do work, so it costs a small amount, but provides reliable data about the state of relays the client has sessions with.

4. **Relaying messages to decentralised network services** - because all relays on Indranet, and even clients, must have a Lightning server and thus a full or light (Neutrino SPV node) Bitcoin node, providing anonymised messaging to these protocols is the first type of exit traffic. Integration with other decentralised services will proceed later, and enable uploading or accessing content on networks like IPFS and Bittorrent, and eventually also to include decentralised messaging, social network, software/media repository hosting, and any other network, including federated and semi-decentralised systems like "cryptocurrencies".

## Protocol Concepts

### Packet and Message Encryption

Indranet uses a message encryption scheme based on [Elliptic Curve Diffie Hellman](https://en.wikipedia.org/wiki/Elliptic-curve_Diffie%E2%80%93Hellman) (ECDH) key exchange.

The message and packet headers contain the following elements:

- **Message checksum** - 4 bytes of the truncated hash of the remainder of the message or packet, for preventing tampering and ensuring integrity of the message.
- **Initialisation Vector** - cryptographically secure random value used for the payload encryption.
- **Cloaked public key** - generated via the use of a strongly random 3 byte value that is concatenated with the receiver's public key, and the first 5 bytes of the combined hash is concatenated to the 3 byte nonce value to prevent inferring association of a stream of message packets with each other. This key also acts as a session identifier, and must be cloaked in order to not provide information to malicious nodes who would then be able to correlate messages.
- **Signature** - The hash of the payload data is signed with a private key that formed the other half of the ECDH enabling securely conveying the encryption secret to the intended receiver over an insecure network. These keys are generated using an algorithm that efficiently generates secure private keys that are correctly on the [secp256k1](https://en.bitcoin.it/wiki/Secp256k1) elliptic curve.

### Signing/Encryption Key Generation and Message Segmentation

The signatures on messages must be different for each subsequent message, and when a message exceeds 1382 bytes (based on a 1410 byte MTU, typical for mobile networks) the message will be segmented into pieces of this size, the last packet padded out with random based hash chain generated noise.

These signing keys are generated by creating two secure, secp256k1 private keys, the base and the secondary, and the base is scalar summed with the secondary to produce a new key, and this new key is then used again the same way for subsequent keys.

This scalar sum operation guarantees that the new private key is also a valid secp256k1 curve point (values not in the curve weaken the encryption), and can be performed very quickly without the resultant key being outside of the curve. Deriving the public key takes approximately the same time but a signature also maps to a message hash so it acts as a MAC (message authentication code) as well as providing half the encryption key to the receiver.

This scheme helps guarantee that once messages are encrypted, even if an attacker gains access to the network handler packet cache they cannot access the encrypted payload data.

### Onion Path Topology

In Indranet there are two primary types of messages, one having a hexagonal shape, the other being diamond shaped. Here is a visual representation with a legend indicating message types in each:

![](onions.svg)

### Proxy

**Proxy** messages are the standard for messages where the **client** is sending messages through a proxy, called the **exit**. Each of the colours shown in the diagram represents the message type. 

- **Forward** messages are purely constructed by the **client**. They are to be carried forwards to a specified IP address, which will be the next hop in the path, or an **exit**, and use the **forward** key provided by relays on purchase. The next hop will decrypt the payload, which can contain either another **forward** or an **exit** message.
- **Return** messages consist of a preformed header encrypted to three secondary keys provided in the purchase process, derived from key pairs embedded in the encrypted layers of the header, called "return" keys, and three ciphers that are to be used sequentially to encrypt the reply message, and the cloaked address in a **return** relay message header layer combined with a public key the client generates to secure the message.

	The provided ciphers for a layer uses the second key known to the **return** relay and client but not the **exit** relay, thus they are unable to unwrap the encrypted message giving the address to the next hop, but by the header and a type byte in the message the **return** hop node then switches to the secondary key that the **exit** doesn't know, that corresponds to the cloaked key, which generates the cipher the **exit** hop was given, combined with the public key visible in the first header. (This is based on the reply messages in Sphinx).

	The second **return** hop also uses this alternate key in order to conceal to the relay whether it is first from the exit or second, the last header only has a packet identifier, because the cipher for the payload is cached by the sender alongside the identifier. This slightly elaborate scheme enables pure forward source routing with the only caveat that exits know they are exits, but **forward** relays don't know whether they are receiving from client, or first, or second hop, and likewise **return** relays can't know whether they are receiving from an exit or sending to a client.

	When a client lacks the **forward**/**return** key pair from a session purchase, it uses a 5 hop circular forward chain which delivers an identifier and the three required ciphers, which the relay will cache for a time expecting a purchase message with which these keys will be used to create the **return** onion encryption.
- **Exit** messages are a special type of message. They contain three symmetric cipher keys that must be used in sequence over the payload, and then appended to the preformed routing header with the return byte flag set.

### Ping

**Ping** messages are purely network probes, used to gauge the reliability of relays the **client** has sessions for. They are constructed out of 3 part onion layers, which ensures that they appear the same as the first two hops in a proxy message to any relay.

Clients perform **ping** messages periodically to check the state of nodes they might use, proactively, to derive several metrics to evaluate their use as hops in **proxy** onions.

### Client

Unlike Tor and other anonymising protocols, every client has the capacity to act as an **exit** for traffic while it is online, for at minimum, Bitcoin and Lightning Network messages. They advertise themselves as "unreliable" exit nodes, this descriptor indicating that they are intermittently offline, and do not attempt to stay online.

This increases the size of the anonymity set for these types of messages, and can include more exit protocols if the user is using them, such as IPFS and other decentralised protocols. It also makes it relatively simple for users to create small, low volume channels for Lightning Network, enabling direct, self-custodial LN payments.

In the diagram above, we distinguish the **client** with blue, but to the nodes before them in the circuit, they appear the same as the one sending, so, **forward** relays see a forward message to **client** and **exit**, and **return** relays see return, and **return** hops see **exit** as return.

All messages look the same as packets in transit, and have no common data between them to establish relationships other than timing. For this reason, the network dispatcher shuffles packets as it sends them out as well.

## Payment for Traffic

Using [Atomic Multi-path Payment](https://docs.lightning.engineering/lightning-network-tools/lnd/amp) (AMP), relays advertise their LN node public keys alongside their Indra identity keys, they send a message constructed as follows:

- Preimage hash of the forward and return private keys that will be delivered after payment to prove payment and provide the keys for handling return messages.
- Amount of bytes the payment is for.

The payment will be for an amount of satoshis in accordance with the rate advertised by the seller. There is no direct return confirmation in this process. As with the rest of Indranet's design, the client is in control of everything, tightening the security.

In the initial bootstrap, the client will send out 5 such payments to establish enough hops to form a secure path. With 5 payments made, relating to 5 sets of **forward/return** keys by the hash of these keys being the preimage used.

This payment delivery onion can deliver 1 to 5 key sets, to handle all possible cases, the only difference between them being the additional key layer in each hop, followed by a forward for the next, the message type is named "SendKeys", which is what follows an LN AMP, and the bundled LND server then relays the bytes amount and preimage to Indra which then holds this to confirm the **SendKeys**.

### Sessions, Depth and Direction

The nodes that are directly adjacent, the one the onion is sent to by the client, and the one that returns a response to a client, can easily associate a session key with the IP address. The second level of depth is less problematic in this way, but has weaker anonymity than the exit hops.

Thus, sessions have a value associated with them to indicate what depth they can be used, and which direction. A simple signed byte integer is sufficient to cover the usual 3 layers of depth. In a ping message there is 1, 2, and -1. In an exit message, there is 1, 2, 3, -2, -1.

This also indirectly brings up the subject of the amount of bandwidth that would be paid for in each case, since each relay needs a session for each of the 5 positions it might be selected to use. Because `AMP` is a fast and relatively cheap operation, it also reduces the risk of loss due to a node failing to deliver service, as the cost paid is low per unit and this observed failure will reduce the chances of the node being selected for future work.

As a relay dutifully performs its service it can then be selected to be used for further sessions at different positions and possibly at larger amounts of bandwidth at a time to reduce the per-transaction cost overhead.

## Relay to Relay Traffic

Messages are segmented into 1382 byte segments and reassembled by relays when they receive them. The relays return an acknowledgement being a signature on the hash of the packet data (which includes the checksum prefix), and these are dispatched in a stream after shuffling by the sending relay, as well as interleaving messages passing to the common next hop when this happens. 

The relay receives the batch of packets, and when it receives enough pieces to reassemble it, according to the error correction data in the packets, and succeeds, reads the message instructions and does as is requested, either a **forward** or **return** message, the difference being described previously. 

### Relay to Relay Encryption

In order to further secure traffic, relays in their chatter with each other provide private relay-to-relay keys to use for message encryption, which are rolled over at least once a day. 

### Dynamic error correction adjustment for Retransmit Avoidance

Based on the conditions of the paths between two relays, by the ratio of packet loss the nodes adjust the error correction to use in order to maintain a margin above the current loss rate, built using a moving average of successful deliveries versus failed.

## Client Path Generation Configuration

A flexible configuration system for selecting paths and exit points is required to cover several different types of use case of obfuscated traffic paths.

- Geo-location based exit and route hop selection:
	- Users may need to avoid using exits within their own or some specified outside jurisdiction.
	- Users may specifically want their exits to emerge in a specified geographical region.
	- Users may want to specify, or avoid selecting intermediate paths in a list of specified geographical regions.
- Selection of specific routers for exits for a given protocol:
	- Using a user's own servers, this can be generalised to allow remote access to a server controlled by the user.
	- A company may provide specific services that users can access at a given set of addresses, whether IP based or domain based.

## Rendezvous Routing for Hidden Services

In much the same way as for emerging directly at a selected router, rendezvous allows relay operators to provide services without revealing their location to the network. This functionality is fairly simple, it is a type of exit that requires the use of a specified public key used by the hidden service operator to negotiate routes to the rendezvous points, the router at the rendezvous essentially takes the packet received at the end point and instead of directly forwarding to another address, sends it in **return** messages to the hidden service's outbound rendezvous path.

This functionality is slated for later implementation after direct, at-router and forwarded connections are implemented, as it requires the design of a protocol for the hidden service to request relaying and wait for **return** messages as they arrive. This is a relatively complex addition to the protocol, but it is important to add it later as this enables both forward **and** backward privacy, whereas the main purpose of the protocol is focused on providing forward privacy, and the exit services are advertised by nodes associated with an address.

## Private Relay Services

To enable users to use Indranet as a transport for accessing private servers, deployments using the `Neutrino` SPV bitcoin node and `lnd` can configure a public key certificate that they can use with a private key, in a similar way to SSH certificate encryption, to enable routing from any client to their specified node identified with its IP address, where there is a relay running at that IP address with the public key registered as enabling access to forward connections to a defined loopback address where the hidden service is running.

This will enable SSH, FTP, and similar services for users to be accessed via Indra, while preventing third parties from identifying the origin of access to the server. This will also enable things like remote desktop access, but it does not include rendezvous routing.

## The Indra Tax

In order to fund development of Indra, the default way that clients open channels to pay for bandwidth as follows:

- All nodes, both clients and relays, by default set up a channel to the Indranet seed node network, which is run by Indra Labs.
- By default, relays must also open relays to a number of other relay nodes, in order to interconnect the Indranet swarm to route payments to relays.
- It is possible for clients to establish other channels if they wish to, and initially this will simply require the understanding of how to control LND's channel protocols, but for the average user, their payments for Indranet bandwidth will always pass through, and usually pass through the LND nodes on Indra Labs seed nodes, enabling a stream of income to fund ongoing development of the system.

This creates an incentive for investors to buy in and operate seed nodes, which decentralises the seed nodes, and we charge them a subscription fee for inclusion in the seed node list, in addition to the gatekeeper routing fees we can charge on our own seed nodes.

-----

# Fin
