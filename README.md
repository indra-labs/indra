[![Indra Routing Protocol Logo](docs/logo.svg)
](https://github.com/orgs/indra-labs/projects/1/views/1)

# Indranet

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/indra-labs/indra)
[![License: CC0](https://img.shields.io/badge/license-CC0-orange.svg)]([http://unlicense.org/](https://creativecommons.org/share-your-work/public-domain/cc0/))

Lightning powered distributed source routed relay network protocol.

[White Paper](docs/whitepaper.md)

## Contributors Welcome

File issues relating to what you want to work on, and we'll add you to the 
team and set up your tasks on the kanban. 

For now the Law is what @l0k18 and 
@lyowhs say it is, but mainly just be nice and keep your inputs within the
critical path towards release.

Code, documentation, the white paper, logo 
designs, whatever floats your boat. If you like and are good with Github Pages
we can get a landing page started.

Even, if you just want to do the journalist thing with us and promote this
project on social media, note that mostly we don't use them, other than 
nostr and twitter. Just make an issue, or contact us through other ways, we
are most willing to do interviews to present our work.

## About

The ubiquitous use of encryption on the internet took some time to happen,
there was a time when the US government defined them as munitions and
claimed export restrictions, and famously the PGP project broke this via the
First Amendment, by literally printing the source code on paper and then
posting it, it became recognised that code, and encryption, are protected
speech.

With ubiquitous 128 and 256-bit AES encryption now in use by default, the
content of messages is secure. However, the volume of messages and endpoints of
signals are still useful intelligence data, enabling state level actors to
attack internet users and violate their privacy and threaten their safety.

Protecting against this high level attack the main network currently doing
this work is the [Tor network](https://torproject.org). However, this system
has many flaws, and in recent times its centralised relay registry has come
under sustained attack by DDoS (distributed denial of service) attacks.

More specifically, the protocol has a severe bottleneck in its rendezvous model for linking two outbound 3 hop connections, attackers flood these with requests, and legitimate users cannot get a word in edgewise. Ironically they built a proof of work protocol to give users a way to get ahead of the spammers.

Indra eliminates this problem by using a constantly changing set of introducers and the actual bidirectional anonymity is done by the parties themselves via the source routing headers plus pairs of hops added in front of the recipient's routing header.

One of the big problems of the Tor network is its weak network
effect. There is no incentive for anyone to run nodes on the network, and
worse, the most common use case is tunneling back out of the network to
anonymize location, is largely abused and led to a lot of automated block
systems arising on many internet services to prevent this abuse.

Indra makes it possible for anyone to offer this kind of outbound relaying service if they want to, but with compensation for doing so, which covers the risk they take as being the visible origin point of shady traffic from time to time.

Indra uses source routing, similar to the Lightning Network and an early but not really quite viable mixnet design called HORNET. The problem with source routed mixnets is that they are very vulnerable to spam. Indra eliminates this problem by no traffic being relayed without first paying a small forward payment to relays for this traffic, thus creating an economic disincentive for spam if the profit is below the routing fees.

## License CC0

![http://creativecommons.org/publicdomain/zero/1.0/](http://i.creativecommons.org/p/zero/1.0/88x31.png)
