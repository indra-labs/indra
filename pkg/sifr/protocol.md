# sifr protocol specification

All nodes advertise a public key for which they have a private key that can be used to send secure messages to them with a shared secret created by Elliptic Curve Diffie-Hellman key exchange.

Sending initial message:

- Select recipient and acquire publicly advertised public key - RPub1
- To: field with RPub1 fingerprint
- Generate new private key for encryption - SPriv1
- Signed: Sign cleartext message with SPriv1
- From: field with SPub1 from SPriv1
- Generate cipher1 from RPub1 and SPriv1
- Nonce: generate new nonce
- Message: encrypted by Cipher1

Receiver receives message:

- Recognise fingerprint to match RPub1
- Combine message Spub1 with RPriv1 to get cipher.
- Decrypt message

Receiver needs to send message back based on message received:

- Receiver public key becomes the public key in the ECDH cipher generation: RPub2
- Generate new private key for reply message: SPriv2
- Sign message with SPriv2
- Generate public key from private key: SPub2
- Generate new message Cipher2
- Encrypt reply message with Cipher2
- To: field with fingerprint of RPub2
- 
