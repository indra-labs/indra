# Indranet Architecture Specification

This will be a top down view of the processes and the data that they manipulate in order to perform the tasks required to implement the protocol.

## Relays

- find the network using DNS seeding
- share addresses of routers in the p2p network with each other

##  Clients

- find the network using DNS seeding
- also share addresses of routers with routers, but send them in order of reliability, creating a collaborative ranking consensus of most popular to least popular
- acquire a collection of sessions with Relays that permit them to pay for traffic.
	- Purchases are always done via onion proxies:
		- send message with onion for payment request wrapped in onions, and a total fee which includes the fees for each hop, plus a noise factor
		- node accepts payment, deducts its fee, forwards request to next in the onion, pays the fee
		- after 3 hops the seller is reached, who takes their payment
		- each hop has a return routing session open waiting for the return path, passing composite cipher encrypted token data back to buyer
	- Onion structure for payment routing
		- Outer layer encrypted to first hop, contains return session cipher which will be used by next hop on the return trip
		- Fee specification, matching what relay advertises for its purchase fees
	- Relay receives onion, replies with Lightning Network invoice for payment, previous hop pays invoice, and when paid, forwards payment onion on to specified next hop, holding the return payment identifier and encryption cipher to be used.
	- When final hop is reached, all 3 hops have open payment session ciphers and identifiers waiting for the return message, which is forwarded back as pre-arranged.
- Once client has at least 6 sessions purchased, it can create circuits to perform proxy relaying
- Each circuit consists of 5 onion layers, the 3rd layer is the exit and the remaining two are return path hops.
- The circuit is circular, as contrasted to Tor circuits, and the circuit is constructed in one step by the client. 
- Similar to the payment circuits, on the return two hops the nodes create open relaying sessions with identifiers tied to the paid for session, and when the exit point receives a reply related to the message sent out, it is passed back through the return hops.
- When the client gets a timeout and clearly a node in the circuit has failed or is congested, there is a probing message type that doesn't consume any bandwidth allocation if it doesn't happen frequently, which consists of a set of two return hops encoded into onions for each step on the path, and the one that fails to return is designated as faulty and a new relay is chosen for this hop in the path.
- An option that can be set on for circuits is to run a probe circuit test for a small extra cost of bandwidth allocation, which can be used to shorten the response time to a dead path and identify the node by sending out these trace path onions regularly, between 1 and 10 seconds depending on the application's liveness requirements. This enables the use of the network for time sensitive applications such as VR and multiplayer game environments, where more than a 1 second delay is expensive to the user and worth the extra bandwidth allocation cost. These extra fees are not greatly expensive, as the probes only contain simple relaying information in addition to the standard header for each onion, but they cost more for giving the ability to rapidly reroute circuits when they fail.
- In addition to the 3 hop out-and-back for payments and the 5 hop round trip circuit pattern, there will be the possibility of customisable relay paths such as concurrent parallel circuits of shorter length (such as only one intermediate, or 3 hop circuits) that provide stronger latency guarantees as well as increasing the network's overall signalling randomness.
- Relays splice their relay traffic together between the many sessions they have running in a random way so as to eliminate easy in-out observation to establish timing relations that signify a circuit.
- Messages are broken into standard network packet size of 1472 bytes with the payloads being split in some cases. The address, or 'to' cipher of the packet header identifies the session via a cloaking mechanism that uses secure random bytes hashed with the public key for receiving messages on this session, eliminating possible correlation.
- The return address key is provided within the encrypted payload for point to point messages with call/return messaging patterns such as ping/pong and question/answer queries.
- Every split 1472 byte long packet is encrypted with a shared secret generated from the cloaked 'to' key and uses two private keys, which the second is summed repeatedly for each subsequent packet before being used to sign it and provide the public part for the receiver to acquire the cipher. This further ensures that there is no common pattern to the data as no data relating to the difference between the ciphers is generated while allowing more efficient derivation of secret keys to be used to generate the signatures that hold the public key the receiver needs to decrypt, in combination with the private key corresponding to the cloaked public key.
- To improve accessibility of the system and avoid relay runners needing to run nodes on remote VPS, Indra Labs will offer a low cost point of presence service, probably around $5/month for a typical home connection speed (up to 100mbit) which allows relay runners to have their system physically accessible. In addition, by using Neutrino as the Bitcoin node in combination with this there will be the possibility of a lot more users running routers from their home connections, further increasing the size and anonymity of the network.