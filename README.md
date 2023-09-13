[![Indranet Logo](docs/logo.svg)
](https://github.com/orgs/indra-labs/projects/1/views/1)

# Indranet

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/git.indra-labs.org/dev/ind)
[![License: CC0](https://img.shields.io/badge/license-CC0-orange.svg)]([http://unlicense.org/](https://creativecommons.org/share-your-work/public-domain/cc0/))

Lightning powered distributed source routed relay network protocol.

[White Paper](docs/whitepaper.md)

## About the Indra Network Protocol

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

More specifically, the protocol has a severe bottleneck in its rendezvous model
for linking two outbound 3 hop connections, attackers flood these with requests,
and legitimate users cannot get a word in edgewise. Ironically they built a
proof of work protocol to give users a way to get ahead of the spammers.

Indra eliminates this problem by using a constantly changing set of introducers
and the actual bidirectional anonymity is done by the parties themselves via the
source routing headers plus pairs of hops added in front of the recipient's
routing header.

One of the big problems of the Tor network is its weak network
effect. There is no incentive for anyone to run nodes on the network, and
worse, the most common use case is tunneling back out of the network to
anonymize location, is largely abused and led to a lot of automated block
systems arising on many internet services to prevent this abuse.

Indra makes it possible for anyone to offer this kind of outbound relaying
service if they want to, but with compensation for doing so, which covers the
risk they take as being the visible origin point of shady traffic from time to
time. We are not going to expressly promote this as we believe that the service
provider is the sole authority over their services, subject to their fidelity to
providing them honestly and justly.

Indra uses source routing, similar to the Lightning Network and an early but not
really quite viable mixnet design called HORNET. The problem with source routed
mixnets is that they are very vulnerable to spam. Indra eliminates this problem
by no traffic being relayed without first paying a small forward payment to
relays for this traffic, thus creating an economic disincentive for spam if the
profit is below the routing fees.

## License CC0

![http://creativecommons.org/publicdomain/zero/1.0/](http://i.creativecommons.org/p/zero/1.0/88x31.png)

## Mission Statement

We are building Indra to end the information asymmetry between large and small
entities on the internet.

We want to earn something from doing this and so we intend to act as introducers
in the initial few years of operation of the main network.

Once the network is running smoothly, we want to distribute that introduction
privilege as widely as possible, at first to several prominent, trusted
organisations, who are diverse and far flung from each other and ideally, in
competition with each other, so that the users interests are paramount.

Regarding copyright and patent, we strongly disagree with any form of monopoly,
no matter what. History shows over and over again that monopolies are fragile
and do not help the people, only the cabal who organise it.

But where the monpoly courts rule in favour of such claims we simply withdraw
from such a lawless place.

We are building Indra because the little people who don't have thugs with titles
to defend fiefdoms, will have far greater freedom to communicate and organise
when the thugs don't have a bird's eye view of the network.

So there is no need to defend such rights since the mission matters more. Plus,
traveling is great for getting new ideas seeing how other cultures solve
problems and the philosophies that lead to their selection.

Indra's anonymous communication system really will form an "outside"
jurisdiction everywhere. Like bringing Lex Merchant onshore. Both the founders
of this project are very interested in laws of contract and tort and equity, in
addition to our CS background. As such, it will enable far more things than you
can list in hours of brainstorming.

The root of power is the control of the flow of information. An enemy cannot
prevail over an army if it never knows the internal state of the enemy, that all
attempts to gather intelligence fail to be either actionable or effective.

The Byzantine Generals Problem was solved for the broadcast case as with
Bitcoin, and many altcoin projects seek to achieve this via federations, which
in our opinion due to economic and game theory reasons is doomed to failure.
Indra builds the solution for the peer to peer case, and we acknowledge that we
are building on the platform that the Bitcoin and Lightning Network community
has built.

We are using the same principles as used in Lightning, and indeed even,
Lightning network developers are unwisely stepping into trouble attempting to
enable their payment channels to carry arbitrary messages. To our minds, this is
as foolish as adding a complex, Turing complete language to the broadcast model
distributed ledger, it is just going to end up proving that when you make
obnoxious behaviour cheap it proliferates.

We intend to be the answer to the question that led to this perilous situation.

Indra will be more than just a way to enforce your privacy, the micropayment
monetised relay service model is wide open to all kinds of innovation, and as
such, it is a core goal that we wish to achieve is a system that enables users
to generate their own message protocols that are executed by Indra relays.

Single point relaying, like is the standard for VPN service, is possible with
Indra's design. So is absurdly long paths, and forking, joining, and deliberate
relaying delays for asynchronous messaging systems all will be easy for
developers to specify in their Indra-savvy applications.
