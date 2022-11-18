![Indra Routing Protocol Logo](doc/logo.svg)

# Indra

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/Indra-Labs/indra)
[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)
[![](https://img.shields.io/badge/chat-telegram-blue)](https://t.me/indranet)

Lightning powered distributed virtual private network with Bitcoin and Lightning
integration.

[White Paper](doc/whitepaper.md)

[Protocol Specification](doc/protocol.md)

## About

The ubiquitous use of encryption on the internet took some time to happen,
there was a time when the US government defined them as munitions and
claimed export restrictions, and famously the PGP project broke this via the
First Amendment, by literally printing the source code on paper and then
posting it, it became recognised that code, and encryption, are protected
speech.

With ubiquitous 128 bit AES encryption now in use by default, the content of
messages is secure. However, the volume and endpoints of signals are still
useful intelligence data, enabling state level actors to attack internet
users and violate their privacy and threaten their safety.

Protecting against this high level attack the main network currently doing
this work is the [Tor network](https://torproject.org). However, this system
has many flaws, and in recent times its centralised node registry has come
under sustained attack by DDoS (distributed denial of service) attacks.

One of the big problems that I saw with this network is its weak network
effect. There is no incentive for anyone to run nodes on the network, and
worse, the most common use case is tunneling back out of the network to
anonymize location, is largely abused and led to a lot of automated block
systems arising on many internet services to prevent this abuse.

The use case that Indranet is first targeted at is protecting location
origin data for Bitcoin transactions and Lightning Network channels. The
increasing value of the currency makes it potentially profitable for the
harvesting of geolocation data associated with targets in order to
physically attack them and take their bitcoins. There has been more than a
few such incidents already, and this is likely to trend upwards and make the
Tor network an ongoing target to stop these transactions from working and/or
unmask their locations and enable further escalation.

Lightning, in particular, currently half of the network capacity is routed
through nodes running on Google Cloud and Amazon Web Services, forming a
very large soft point for governments to harm the routing capacity of the
network, impeding adoption, and potentially making a way for users to be
robbed by state sized actors like the CIA, FSB, MI6 and similar
organisations with zero accountability.

Thus, Indranet's main task is in fact creating a network of hidden services
that are used by Bitcoin and Lightning node operators to perform
transactions that will not be detectable or locatable by even large scale
actors.

Thus, it is essential that routers on Indranet get paid for their work, in
order to maintain their connection and equipment costs.

Clients purchase session keys via a circuit of proxies using onion encoded packets
and lightning payments which can then be used to spend bytes from their sessions in
arbitrary, source routed paths exiting to protocols which are also deducted
from the session bandwidth allocation.

In this way, nodes are unable to correlate between payments through LN and
the spending of their vouchers, allowing routers to be paid, and thus
incentive to increase routing capacity through the ability to then pay for
the infrastructure running the network.

Indranet will use a programmable, client side onion construction
scheme that will be designed to be configurable and programmable such that
the uniform three hop pattern can be extended to include parallel multipath and
dancing paths and make tradeoffs for latency, reliability and obfuscation
for other purposes. In addition, it forms a universal routing layer that
enables users to get around the currently complex and sometimes impossible
restrictions on inbound traffic caused by IPv4 and Network Address Translation.
