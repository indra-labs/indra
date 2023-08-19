# Indranet Protocol Message Format

In order to fully discover the ways in which data structures can be composed, it is important to elaborate the parts, and the many common patterns that these parts can be composed into, to ensure the part set is complete and sufficient to implement an open ended system that can adapt to uses that were not envisioned initially.

Indra is essentially a distributed computation system whose main task is relaying messages via compositions of layers of messages that are progressively unwrapped as they pass through the relays the client designated them to pass. Each message type functions like an API call, the most frequent and core instruction being to relay a message to another node.

## Top Level Message Types

### Session

The session is the most important and primary message type in the sense that it must be delivered in order for a relay to be obliged to perform services for clients.

The session message contains two symmetric encryption keys, one called the Header key and the other called the Payload key.

As described elsewhere, this is to enable the construction of a pre-configured message encryption header which provides the instructions to place in front of a message, and the encryption keys to add the layers of encryption that contain the knowledge of the message pathway to the immediate origin and next destination, while making it very difficult to see any further than this without controlling all of the relays specified in the path.

The session message is the only message type that Indranet relays will process a subsequent forward message without requiring a session to already exist, under the proviso that there is a received payment for which the preimage hash matches the hash of the pair of symmetric secret session keys that the client will use in the encryption of messages that will be forwarded by the relay.

At the same time serving to both secure the message to be only readable by the authorised relay, and identifying the message so the session can be billed for.

Sessions contain a **Header** key and a **Payload** key, which is described in the **Reply** section in the following.

The session is a reference to a pre-paid balance, against which a bytes/time rate is applied to messages that are forwarded via the session. The session also can be alternatively billed on a different rate for **Exit** messages, as described below.

### Forward

The number one task of Indranet relays is to accept a message, and forward it to another relay.

The Indra protocol is [connectionless](https://en.wikipedia.org/wiki/Connectionless_communication) because relays do not participate in making routing decisions.

However, because it is necessary to enable arbitrary delay instructions, and because it can happen that clients are out of date with the state of the network, and such problems as congestion, network and software failure, and the changing of IP addresses, the forward needs to contain one primary data element, which is the identity public key of the relay.

Relays must keep a database of metadata about relays that provides them with a mapping between these public identity keys and the current IP addresses that can be used to reach it.

Thus, the format can be depicted as follows:

| 1 | Forward message magic bytes (4 bytes human-readable sentinel) |
| ----- | ------------------------------------------------------------ |
| 2 | Relay Identity Public Key (33 bytes)                         |

### Crypt

The second most important message type in Indranet is the **Crypt**.

The crypt is an encrypted message, consisting of a header containing a cloaked session key referring to the session private key, and the random seed value used to prevent the possibility of plaintext cryptanalysis attacks.

The crypt specifies encryption that is used to "wrap" the remainder of a message so that only the intended recipient can see it, a combination of encryption and authentication rolled into one.
| 1 | Crypt message magic bytes  |
| ----- | ------------------------------------------------------------ |
| 2 | *Cloaked* * session **Header** public key (8 bytes, 4 bytes blinding factor, 4 bytes truncated hash of public key and blinding factor) |
| 3 | Message public key (sender generated one-time)                                    |
| 4 | Initialisation Vector (16 bytes standard AES high entropy random value) |

>  \* Cloaked means concealing the session key by taking a 4 byte random value, the **Blinding** factor, concatenating the public key after it, hashing the concatenated string, and then truncating the hash to 4 bytes, and concatenating it to the random value. 
>
> The relay can then scan its session database by generating the same construction using the same method just described to determine if the candidate key matches.
>
> The reason for this is to prevent the relay from correlating two packets that it may be forwarding to the same next hop relay or client, as being related via the **Session**.

### Reply

In order to facilitate the return of an arbitrary blob of data as a reply to a message sent out by a client, there is a special construction of pre-made message which contains an header containing the forwarding and encrypted layers for the reply message.

The general design of the **Reply** message is a **Forward** message, designating the intended that will perform the forwarding, and the use of the **Header** key of the pair of keys to generate a layer of **Crypt** that wraps the subsequent message layer, and this is repeated an arbitrary number of times.

> In order to prevent the depth of the chain of forwards from being visible to relays, there must also be a random, arbitrary padding at the end of the header. Initially a rigid design was intended to cloak this, hiding the position on the path by it being moved upwards and padded back out for the next step, so a random length of padding that varies enough to make it difficult to know how many layers might be inside it must be used.
>
> Because, also, the size of the Forward and Crypt messages are fixed, this header will be padded out as though there is one or several more layers than are actually present, in order to obscure any information about the real length of the path.

#### Header

| 1 | Forward |
| ----- | ------------------------------------------------------------ |
| 2 | Crypt |
| 3 | ... repeat 0 or more **Forward**, **Crypt** layers |

##### Extra Data

| 4 | Ciphers |
| ----- | ------------------------------------------------------------ |
| 5 | Initialisation Vectors |
| 6 | Pad |
| 7 | Sentinel |

#### Ciphers

Sessions consist of two symmetric secret keys, the **Header** and the **Payload** keys. The **Header** key is used in the header shown above, to derive the Cloaked session header public key. 

The Ciphers contain a series of symmetric secret keys that are the product of using ECDH on the one-time public key in the **Crypt**, and the session **Payload** key. The relay can thus encrypt an arbitrary message payload using this key

#### Initialisation Vectors

The IVs used with the ciphers above, and wrapped in the **Crypt** messages, must be a separate set of IVs from the ones in the header. They must also be the same number as the **Ciphers**.

#### Pad

In order to prevent the intentional tagging of packets with an arbitrary size, the amount of padding to be added to the reply payload is provided by the client. In this way, any attempt to pad the message arbitrarily can be detected, and correctly estimated to be an attempt to make the reply identifiable to co-conspirator relays elsewhere in the network to allow the tracking of the message transit.

By making this length under the control of the client, it makes the billing rate for the reply partially under the control of the client, who must pay for this extra data.

#### Sentinel

A special, random string of bytes is provided alongside the pad length that is to be placed at the end of the padding, which when the reply is received, the client can see where the original message ending was, and thus the pad, which if it is a mismatch to what was prescribed, is a bannable offense that will cause the client to deprecate the use of the exit node as a consequence.

If this recurs, the next ban score increase must be a multiple of the previous, until the threshold for outright permanent banning shall be applied. This should only take at most 3 incidents to be certain there was no accident.

## Exit

The Exit message is a request to tunnel a packet out of the Indra network.

Indra nodes advertise these as *Services*, which are identified by a Well Known Port, such as 80 for HTTP, 443 for HTTPS, 25 for SMTP, 22 for SSH, and so on.

Traffic that is forwarded to a *Service* is billed according to the average of the inbound message size and the outbound message size that is received back from the service in response to the request that was received inside the Exit message.

