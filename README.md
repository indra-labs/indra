![Indra Routing Protocol Logo](docs/logo.png)

# Indranet

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/indra-labs/indra)
[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)
[![](https://img.shields.io/badge/chat-telegram-blue)](https://t.me/indranet)

Lightning powered distributed virtual private network for anonymising traffic on
decentralised protocol networks.

[White Paper](docs/whitepaper.md)

## About

The ubiquitous use of encryption on the internet took some time to happen,
there was a time when the US government defined them as munitions and
claimed export restrictions, and famously the PGP project broke this via the
First Amendment, by literally printing the source code on paper and then
posting it, it became recognised that code, and encryption, are protected
speech.

With ubiquitous 128 and 256 bit AES encryption now in use by default, the
content of messages is secure. However, the volume of messages and endpoints of
signals are still useful intelligence data, enabling state level actors to
attack internet users and violate their privacy and threaten their safety.

Protecting against this high level attack the main network currently doing
this work is the [Tor network](https://torproject.org). However, this system
has many flaws, and in recent times its centralised relay registry has come
under sustained attack by DDoS (distributed denial of service) attacks.

One of the big problems that with this network is its weak network
effect. There is no incentive for anyone to run nodes on the network, and
worse, the most common use case is tunneling back out of the network to
anonymize location, is largely abused and led to a lot of automated block
systems arising on many internet services to prevent this abuse.

# fin

notes:

`([a-zA-z0-9\_\-\.][a-zA-z0-9\/\_\-\.]+)\:([0-9]+)` is a regex that matches the
relative file paths in the log output. $1 and $2 from this are the relative path 
and the line number.