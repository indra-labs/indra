![Indra Routing Protocol Logo](doc/logo.png)

# Indranet

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/Indra-Labs/indra)
[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)
[![](https://img.shields.io/badge/chat-telegram-blue)](https://t.me/indranet)
[![](https://img.shields.io/badge/chat-keet-blue)](punch://jc38t9nr7fasay4nqfxwfaawywfd3y14krnsitj67ymoubiezqdy/ykm4nji4y6nnnt9u9t89ny1xecizm1aex1k16aoewr31cxsdrbiwaprbhjuzfgpok68mjxhwnark3daxq97cjfcy84rqsk9gpippp44yyry965ncp17akyeyybyet64d9ou73icercidjz4sfxa8i6pedsj467dno4r6h87hpga15pog1r5wcs9enpr7qrnt9zyh7ontjcdukca1kffwuqf5nk9a8egwba)

Lightning powered distributed virtual private network for anonymising traffic on decentralised protocol networks.

[White Paper](doc/whitepaper.md)

## About

The ubiquitous use of encryption on the internet took some time to happen,
there was a time when the US government defined them as munitions and
claimed export restrictions, and famously the PGP project broke this via the
First Amendment, by literally printing the source code on paper and then
posting it, it became recognised that code, and encryption, are protected
speech.

With ubiquitous 128 and 256 bit AES encryption now in use by default, the content of
messages is secure. However, the volume of messages and endpoints of signals are still
useful intelligence data, enabling state level actors to attack internet
users and violate their privacy and threaten their safety.

Protecting against this high level attack the main network currently doing
this work is the [Tor network](https://torproject.org). However, this system
has many flaws, and in recent times its centralised relay registry has come
under sustained attack by DDoS (distributed denial of service) attacks.

One of the big problems that with this network is its weak network
effect. There is no incentive for anyone to run nodes on the network, and
worse, the most common use case is tunneling back out of the network to
anonymize location, is largely abused and led to a lot of automated block
systems arising on many internet services to prevent this abuse.

Indranet does not set itself up to be a direct competitor for the Tor network. In its first few years of operation it will not have any mechanism for tunneling out of the network, and if it ever does, this will be user-contributed functionality, and not encouraged since any node providing exit becomes a target for retaliation when used to abuse such external systems.

Indranet's purpose is to form an interconnect layer for decentralised network protocols. It requires Lightning Network, as a primary requirement, to enable the payment for bandwidth, and thus it also requires connectivity to the Bitcoin network. Thus, all nodes, both relays and clients, will provide exit traffic for these two protocols, especially Bitcoin, which has a very low bandwidth requirement for simple transaction publishing.

Users will potentially be able to set up arbitrary exit services, but the core project will only target connectivity with decentralised peer to peer services. Secondarily, it will be possible for users to set up private, non advertised exit services, protected via certificate authentication, such as SSH and other remote access systems. This will serve to displace the use cases for Tor with SSH and web services.

Later, rendezvous access protocols will be added and enable the creation of arbitrary hidden service addresses such as web applications.

# fin
