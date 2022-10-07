# sifr protocol specification

All nodes advertise a public key for which they have a private key that can be used to send secure messages to them with a shared secret created by Elliptic Curve Diffie Hellman key exchange.

The sender always generates a new private key for each session, and forwards the public key at the head of their message.

