// Package sifr is the general purpose crypto-system for encrypting and
// decrypting messages in Indra.
//
// Senders always generate a new secret key, and attach the public key
// corresponding to it at the front of the message.
//
// This is used by the receiver in combination with the private key
// corresponding to public key they advertise to enable ECDH forward secrecy.
// It ensures that the secret is never the same except for the receiver's public
// key.
package sifr
