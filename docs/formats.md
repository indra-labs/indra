# Indranet Protocol Message Format

In order to fully discover the ways in which data structures can be composed, it is important to elaborate the parts, and the many common patterns that these parts can be composed into, to ensure the part set is complete and sufficient to implement an open ended system that can adapt to uses that were not envisioned initially.

Indra is essentially a distributed computation system whose main task is relaying messages via compositions of layers of messages that are progressively unwrapped as they pass through the relays the client designated them to pass. Each message type functions like an API call, the most frequent and core instruction being to relay a message to another node.

Indranet's messages function in a similar manner to instructions in a scripting language. They are processed in the narrow context of what parts of the message are visible to the relay, and sequential segments that can be seen by a relay are processed via internal processing and pass forward a look-behind in order to enable the relay to process a message in context of something else. Crypts, for example, are the way that session accounting is done, so messages that come after them can use this information to make decisions.

TODO: perhaps the look-behind buffer should be as long as the cleartext messages, with chains of various message types is found, so that ordering is not so important as set associativity.

## Primitives

### Session

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 32 | Header | Symmetric ECDH private key used in reply headers and simple forwards |
| 32 | Payload | Symmetric ECDH private key used for encrypting response payloads |

The session is the most important and primary message type in the sense that it must be delivered in order for a relay to be obliged to perform services for clients.

The session message contains two symmetric encryption keys, one called the Header key and the other called the Payload key.

As described elsewhere, this is to enable the construction of a pre-configured message encryption header which provides the instructions to place in front of a message, and the encryption keys to add the layers of encryption that contain the knowledge of the message pathway to the immediate origin and next destination, while making it very difficult to see any further than this without controlling all of the relays specified in the path.

The session message is the only message type that Indranet relays will process a subsequent forward message without requiring a session to already exist, under the proviso that there is a received payment for which the preimage hash matches the hash of the pair of symmetric secret session keys that the client will use in the encryption of messages that will be forwarded by the relay.

At the same time serving to both secure the message to be only readable by the authorised relay, and identifying the message so the session can be billed for.

Sessions contain a **Header** key and a **Payload** key, which is described in the **Reply** section in the following.

The session is a reference to a pre-paid balance, against which a bytes/time rate is applied to messages that are forwarded via the session. The session also can be alternatively billed on a different rate for **Exit** messages, as described below.

These keys are not straight symmetric ciphers, they must be combined with a public key using ECDH, and for encryption the sender uses the session public key and generates a new private key for each message.

### Forward

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 33 | RelayID | Public identity key of relay to forward this message to |

The number one task of Indranet relays is to accept a message, and forward it to another relay.

The Indra protocol is [connectionless](https://en.wikipedia.org/wiki/Connectionless_communication) because relays do not participate in making routing decisions.

However, because it is necessary to enable arbitrary delay instructions, and because it can happen that clients are out of date with the state of the network, and such problems as congestion, network and software failure, and the changing of IP addresses, the forward needs to contain one primary data element, which is the identity public key of the relay.

Relays must keep a database of metadata about relays that provides them with a mapping between these public identity keys and the current IP addresses that can be used to reach it.

### Crypt

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | Cloaked\* Header public key |Indicates to the valid receiver which public key is being used.|
| 33 | Message public key | (sender generated one-time) |
| 16 | Initialisation Vector | standard AES high entropy random value) |
| 3 | Offset *** | 24 bit vector (up to 16Mb header) for beginning of payload (using Payload key from [Session](#session)) |

>  \* **Cloaked** means concealing the session key by taking a 4 byte random value, the **Blinding** factor, concatenating the public key after it, hashing the concatenated string, and then truncating the hash to 4 bytes, and concatenating it to the random value. 
>
>  fingerprint = hash ( nonce | public key ) -> truncated to 4 bytes
>
>  cloak = nonce | fingerprint
>
>  The relay can then scan its session database by generating the same construction using the same method just described to determine if the candidate key matches.
>
>  The reason for this is to prevent the relay from correlating two packets that it may be forwarding to the same next hop relay or client, as being related via the **Session**.
>
>  ***  The Offset is encrypted as the first 3 bytes of the message, concealing from casual observation how deep the header is relative to the packet.

The second most important message type in Indranet is the **Crypt**.

The crypt is an encrypted message, consisting of a header containing a cloaked session key referring to the session private key, and the random seed value used to prevent the possibility of plaintext cryptanalysis attacks.

The crypt specifies encryption that is used to "wrap" the remainder of a message so that only the intended recipient can see it, a combination of encryption and authentication rolled into one.

### Reply

In order to facilitate the return of an arbitrary blob of data as a reply to a message sent out by a client, there is a special construction of pre-made message which contains an header containing the forwarding and encrypted layers for the reply message.

These **Reply** messages are used anywhere the protocol wants to enable a bidirectional path without the receiver of the reply message being able to unmask the location of the sender.

The general design of the **Reply** message is a **Forward** message, designating the intended that will perform the forwarding, and the use of the **Header** key of the pair of keys to generate a layer of **Crypt** that wraps the subsequent message layer, and this is repeated an arbitrary number of times.

> In order to prevent the depth of the chain of forwards from being visible to relays, there must also be a random, arbitrary padding at the end of the header. Initially a rigid design was intended to cloak this, hiding the position on the path by it being moved upwards and padded back out for the next step, so a random length of padding that varies enough to make it difficult to know how many layers might be inside it must be used.
>
> Because, also, the size of the Forward and Crypt messages are fixed, this header will be padded out as though there is one or several more layers than are actually present, in order to obscure any information about the real length of the path.

#### Header

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 2 | HeaderLength | Length of Header, after which Extra Data is found |
| ... | Crypt | N layers of crypts |
| ... | ... | repeat 0 or more **Forward**, **Crypt** and then optionally **Delay** and **Pad** layers |
| ... | | padding with length of a random number of extra **Forward/Crypt** layers |

#### Extra Data

These are found directly appended to the end of the above header

| Byte length | Name | Description |
| -----| ----- | ----- |
| 2 | Layer count | Number of Ciphers/IVs found in the following. |
| 32 | Ciphers | Pre-generated symmetric ciphers created with Payload session key and the same public key found in the matching Header layer (hidden from Exit/Hidden Service via encryption) |
| 16 | Initialisation Vectors | IVs that match the ones used in the header Crypt layers |
| 4 | Sentinel | Magic bytes indicating the sender defined sentinel to place at the proper end of the response bundled using the Reply |

#### Ciphers

Sessions consist of two symmetric secret keys, the **Header** and the **Payload** keys. The **Header** key is used in the header shown above, to derive the Cloaked session header public key. 

The Ciphers contain a series of symmetric secret keys that are the product of using ECDH on the one-time public key in the **Crypt**, and the session **Payload** key. The relay can thus encrypt an arbitrary message payload using this key

#### Initialisation Vectors

The IVs used with the ciphers above, and wrapped in the **Crypt** messages, must be a separate set of IVs from the ones in the header. They must also be the same number as the **Ciphers**.

#### Sentinel

A special, random string of bytes is provided alongside the pad length that is to be placed at the end of the padding, which when the reply is received, the client can see where the original message ending was, and thus the pad, which if it is a mismatch to what was prescribed, is a bannable offense that will cause the client to deprecate the use of the exit node as a consequence.

If this recurs, the next ban score increase must be a multiple of the previous, until the threshold for outright permanent banning shall be applied. This should only take at most 3 incidents to be certain there was no accident.

### Exit

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID |Database key for retrieving pending message|
| 2 | Exit port | Well known port \* |
| ... | Reply | Header, Ciphers and Initialisation Vectors for bundling Response |
| 4 | Request Length | Length of data following that contains request |
| ...  | Request | Message to forward to Service via local port forward |

The Exit message is a request to tunnel a packet out of the Indra network.

Indra nodes advertise these as *Services*, which are identified by a Well Known Port, such as 80 for HTTP, 443 for HTTPS, 25 for SMTP, 22 for SSH, and so on.

Traffic that is forwarded to a *Service* is billed according to the average of the inbound message size and the outbound message size that is received back from the service in response to the request that was received inside the Exit message.

> \* Well known ports are generally defined in unix for the privileged low ports, but also in protocols, such as 8333 for Bitcoin, 9050 for Tor, etc.
>
> A list will be compiled to add any extra features that are not services that would normally be defined as they are not normally public.
>
> For example, proxy port numbers are arbitrary, but we might specify they are to always be, say, 8080 for a HTTP proxy, or 8004 for Socks 4A or 8005 for Socks 5, and so on. These will most likely be the loopback ports that are usually used or even specified in the protocol, even though for clearnet use such an open relay would be a security/spam risk, this is the purpose of Indra, and spam is controlled separately by the metering of data volume for relaying

### Request

| Byte length | Name  | Description                                   |
| ----------- | ----- | --------------------------------------------- |
| 4           | Magic | Sentinel marker for beginning/end of messages |
| 4 | Length | 32 bit value for the length of the message, ie, up to 4 gigabytes |
| ... | Data | The content of the Request |

> todo: maybe it should instead be 24 bits, ie 16 megabytes, more than large enough for a high throughput protocol.

Request is exactly a standard request message for a server as found in all Client/Server protocols. It is simply a wrapper for an arbitrary payload of data.

### Response

| Byte length | Name  | Description                                   |
| ----------- | ----- | --------------------------------------------- |
| 4           | Magic | Sentinel marker for beginning/end of messages |
| 4 | Length | 32 bit value for the length of the message, ie, up to 4 gigabytes |
| ... | Data | The content of the Response |

Response is the message wrapper for the response from a Service.

### Delay

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 4 | Delay | Number of milliseconds to hold message in cache before processing next message layer |

In order to add some temporal jitter and to arbitrarily increase and decrease the size of the layers that are stacked at the head of messages, it is possible to add a specification that provides a millisecond precision 32 bit value specifying an amount of time to hold the message in the buffer before sending it.

In order to facilitate this the relays must charge according to a coefficient multiplied by the time of delays against the message size.

For this reason also, as will be explained with **Pad**, this is entirely under the control of the client for both reasons of securing anonymity as well as permitting applications to add these to messages for whatever purpose the application developer envisions, such as, potentially, storing data temporarily out of band for a prescribed amount of time as a form of "cloud storage".

### Pad (Increment/Decrement)

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 2 | 16 bit value signifying milliseconds of delay |

Outside of the anonymity quality-of-service, reciprocating dummy traffic that peers will send to each other, using a scheme of fractional reserve for allowing temporal disjoint reciprocation, or whatever scheme ends up being used, the client needs to have control over how the size of their messages is altered deliberately in transit.

> (note: talk about this later, eg concurrent sending of same sized message to two forward points, one defined by client, the other randomly chosen)

This is not, obviously, something that necessarily evil nodes are going to abide by, but any that don't, and add arbitrary pad and/or snip off sentinels the client expects will flag themselves as mischief makers, or at least, that mischief happened along the path, putting all nodes in the path under suspicion, for use of automatically detecting anomalous behaviour that repeats and may be abusive.

### Split

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 2 | Count | Indicates the number of splits following |
| 4 | Offsets | *Count* number of 32 bit offset values marking the segments |
| 33 | Destinations | Number of relay identity keys - this is 1 more than *Count* as the boundary of the last split is the end of the message. |
| ... | Data | The segments that begin from the end of the Offsets/Destinations and continue until the end of the message. |

Split is where the remainder of the message has 2 or more segments that each bear a header indicating a different destination. This could be used, for example, to create a liveness detection along an arbitrary route that conceals the return paths of this telemetry.

Splits also make possible the fan-out/fan/in pattern for multipath messages.

#### Low Latency Mixnet "Lightning Bolts"

One of the biggest difficulties with mixnets is that the lower the latency, the easier it is to correlate traffic paths as they flow through the network.

Defeating this attack can be achieved by adding **Split** messages fan out randomly to deliver empty or padding, so that at each hop, at least two different simultaneous transmissions take place.

These can be called Lightning Bolts since they propagate in a similar way as arcs of electricity across the sky and to the ground, forking towards equally conductive or from equally charged areas that merge or split.

The simulation of merging can even be created, as well, with forks that merge back together.

## Network Intelligence 

In source routing systems, the nodes that perform relaying services must advertise their existence and instructions on how to reach them.

All advertisements contain the following 4 fields, and additional fields as required:

### Ad Prototype

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID | Random value that ensures the signature is never placed on the same hash |
| 33 | Key | The public identity key of the relay providing relaying or other service |
| 8 | Expiry | The timestamp after which the ad must be renewed as all peers will evict the record from their network intelligence database |
| 64 | Signature | Schnorr signature that must match with the Key field above |

These are the 4 essential elements in an ad, as shown above, and all the ads for both public and hidden services contain this. The signature, of course, is always at the end, but the order of the fields *could* be different.

### Peer

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID | Random value that ensures the signature is never placed on the same hash |
| 33 | Key | The public identity key of the relay providing relaying or other service |
| 8 | Expiry | The timestamp after which the ad must be renewed as all peers will evict the record from their network intelligence database |
| 4 | RelayRate | The price, in MilliSatoshi, for a megabyte of data |
| 64 | Signature | Schnorr signature that must match with the Key field above |


Peer is simply the advertising of the identity of a peer. It contains only the public identity key, and the relay rate charged by it.

### Addresses

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID | Random value that ensures the signature is never placed on the same hash |
| 33 | Key | The public identity key of the relay providing relaying or other service |
| 8 | Expiry | The timestamp after which the ad must be renewed as all peers will evict the record from their network intelligence database |
| 1 | Count | Number of addresses listed in this, maximum of 256, 0 being the first |
| 8/20/? | Addresses | Network Addresses - variable length. IPv4 and IPv6 encoding lengths with 2 byte port numbers added |
| 64 | Signature | Schnorr signature that must match with the Key field above |


## Services

The first type of service provided over Indranet is public **Services**. These are services that are advertised by relays, that designate routes that messages to them can tunnel out to, outside of Indranet.

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID | Random value that ensures the signature is never placed on the same hash |
| 33 | Key | The public identity key of the relay providing relaying or other service |
| 8 | Expiry | The timestamp after which the ad must be renewed as all peers will evict the record from their network intelligence database |
| 2 | Count | Number of services advertised, consists of Port and RelayRate |
| 2 | Port | 16 bit port number of service, based on Well Known Port |
| 4 | RelayRate | The cost in MilliSatoshi per megabyte of, the mean of request and response byte size |
| 64 | Signature | Schnorr signature that must match with the Key field above |

## Hidden Services

Hidden services are a composition of Primitives that enables the initiation and message cycle of enabling bidirectional location obfuscation.

It borrows a little from the hidden service negotiation protocol used by Tor, but the connection is maintained by the hidden server and hidden client directly, bypassing the bottleneck and small attack surface created by a mediated rendezvous connection.

### Overview

Prior to explaining the parts, it is necessary to list them, and this is best done as a numbered sequence, and the three parties involved we will use the common names used in cryptography that apply to this protocol.

#### Alice

Alice is the hidden service provider. Alice is generally the initiator in most scenarios described in cryptography.

#### Bob

Bob is the hidden client. Bob wants to talk to Alice, but doesn't know where she is picking up messages from.

#### Faith

Faith is often used as a trusted intermediary, however in this protocol she is serving in the role of an introducer, and her service is temporary.

Note that currently there is no scheme for billing hidden service traffic, essentially the cost of relaying is borne equally by the hidden service and the hidden client. Faith is essentially paid for this service as the hidden service must have a session with her to do this. Each time a new introduction is received, she is paid.

> note: in discussions of attacks on this protocol, the name **Eve** would be perfect as the placeholder for the attacker.

### Hidden Service Protocol

1. Alice wants to offer a hidden service, without disclosing to the network the location where the data is being processed or stored. This has especially got relevance to such services as trusted intermediaries, as a repository of secrets, even if encrypted, is a high value target for attackers.
2. Alice generates an introduction message, which consists of her hidden identity key, and the identity key of the chosen intermediary, Faith, is part of this message.
3. The second part of her message is secret, and consists of a **Reply** message, which will be used to forward a request to Alice. In order to prevent  gathering any information about this return path, each one is single use, and after an introduction is consumed, the introducer waits for a new one.
4. The delivery of this introduction, and subsequent new Reply messages is done via an anonymised pathway that is typically 3 hops, but could be shorter or longer. *3 is just the magic number that is the maximum bang per buck for creating a path that can be difficult to trace, similar to how an extraction that takes 50% or better per pass exceeds 90% after 3. This is known as "The Rule of Three".*
5. Faith broadcasts this introduction across the network, in part because when hidden clients request a connection, she gets paid for forwarding the request. 
6. Because Faith has a privileged position as the go-between, where she is doing a one-time version of the Tor hidden service, Alice will rotate the set of introducers, that is, Alice will send out many of these public/private intro/reply messages to peers, thus serving to create a moving target for would-be attackers, and also, on the other side, Faith is the attack surface for attempts to unmask the identity and location of Alice.
7. Bob receives the gossip about the various introducers related to the public identity key of Alice's hidden service, and wants to start a hidden conversation with her, without revealing his location either.
8. Bob sends a request to Faith via an obfuscated multi-hop path, containing a **Reply** header, which is then forwarded back to Alice by Faith using the single use reply header she currently has for Alice.
9. Alice then receives this request, and then constructs a two or more layer forwarding prefix to add an extra two or more hops on top of the path that was given by Bob, the reason being that Bob might control the first hop, and be trying to unmask Alice. In this way, he would be thwarted at such an attempt. This reply is called **Ready**, and is like the Clear To Send signal on a serial connection, this step in the process akin to the handshake of TCP or similar in that now both parties are ready to start a conversation.
10. Bob then receives this message, which contains a Reply header from Alice, and Bob uses this, again with his own header prefix, to send a **Request** message, and wraps this in 2 or more hops to protect his location from the relay Alice provided in her **Reply**.
11. This then gets back to Alice, who can then forward the **Request** on to the actual hidden service server, who then returns a reply and Alice then wraps this in her two or more forward layers, around Bob's **Reply** header and inside that, the **Response** from the hidden service.

### Protocol Messages for Hidden Services

Note that at all times, for reasons of eliminating the possibility of unmasking either end of a hidden connection, each side prefixes their messages with at least two hops in the path. This acts as a firewall against an evil counterparty controlling the relay at the top of the **Reply**.

#### Intro

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | Nonce | Random value that ensures the signature is always on a different hash |
| 33 | Key \* | Public key of the hidden service identity, matching the Signature below |
| 33 | Introducer | Public key of the relay that is serving as the introducer |
| 8 | Expiry | The timestamp after which, by the clock of the introducer that the intro will be evicted from its network intelligence database |
| 64 | Signature | Schnorr signature on the foregoing message that matches **Key** |

The intro is the publicly visible document that contains a signed designation of the introducer (Faith) in the protocol.

It is gossiped over the Publish/Subscribe system of Indra that propagates information about peers, their addresses, and their offered public services.

> \* This key is encoded in the Indranet `Based32` encoding, which is 26 lower case latin letters and 234679, which are the least ambiguous 6 numbers out of the 10 arabic number ciphers. This custom encoding is used because it provides potential for later use of vanity addresses. But more than this... A note is required about this.
> 
#### *Rate Limiting Hidden Service Advertisements*

Because there is no cost to generating, and a relatively low cost to publishing hidden services, in order to limit the amount of new hidden service addresses, they must contain a common 25 bit prefix that forms the word `indra`. Well, this is the provisional idea, it may need to be a longer prefix than this to be sufficiently limited in the context of possible ASIC devices for mining these keys. Currently the public key derivation operation is fairly expensive, very few Tor hidden services have more than 7 or 8 base32 characters.

#### Introduction

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| ... | Intro | The intro advertisement |
| ... | Reply | The Reply message that the introducer returns a Route back with |

Introduction is the message sent out by a hidden service, over a multi-hop path, to the Introducer they have chosen for a time to serve as introducer.

The Intro is gossiped via the Publish Subscribe peer to peer protocol, and every time a client sends a Route request, the hidden service will send a new one, as it will refuse to accept the second one of these. 

This prevents a race condition and the possible plaintext attack that the same cipher set might open up.

#### Route

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID | Random value used by the hidden client to identify the connection request |
| ... | Reply | The path to send back the Ready signal |

Route is essentially a connection request, after which the hidden client will excpect to receive a Ready message containing a new **Reply** header to send a **Request**.

#### Ready

| Byte length | Name | Description |
| ----- | ----- | ----- |
| 4 | Magic | Sentinel marker for beginning/end of messages |
| 8 | ID | The random value that was sent in the request is returned with the reply for quick retrieval |
| ... | Reply | The **Reply** header the hidden client can use to reply |

The ready message essentially functions in the same way as a standard handshake acknowledgement, which establishes that both sides are ready to begin a conversation, and the first request may be sent.

### Hidden Message Cycles

Aside from the unmasking prevention prefixes, and the **Reply** headers, the pattern of **Request** and **Response** are the same as any client/server protocol.

In order to maintain the connection, each side retains a number of prior **Reply** headers (3-5?) that were sent, and in the event of a transmission failure, the former, known successful **Reply** headers are retained for resending the message that did not get a reply. The TTL of such chatter is governed by the protocol that is being transported, on its "Application" layer, to use OSI layer cake nomenclature.

In the event that the client consumes all of its cached past **Reply** headers for the service, it can simply search out, or just use, other **Intro** advertisements that it has received over the gossip network, and reestablish the connection.

> todo: Reply headers probably need an expiry?

## Example Custom Protocols

> todo: some examples of other ways to use the primitives described in the foregoing, here are some headings:
>
> 1. Time Delay Data Storage
> 2. Metered Network Access
> 3. Paywalling access to content/databases/application
> 4. Additionally maybe discuss various strategies for defining paths, most especially, forks. Oh, I will discuss that in a section above this.