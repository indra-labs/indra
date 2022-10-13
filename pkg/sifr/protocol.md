# sifr protocol specification

All nodes advertise a public key for which they have a private key that can be used to send secure messages to them with a shared secret created by Elliptic Curve Diffie-Hellman key exchange.

This table shows the sequence of events from left to right as relates to the fields of the messages:

|            Field | Initiator                                                    | Correspondent                                                | Initiator                                                   | Correspondent                                                |
| ---------------: | :----------------------------------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------------- | ------------------------------------------------------------ |
|               To | Fingerprint of CPub1 (advertised)                            | Fingerprint of IPub1                                         | Fingerprint of CPub2                                        | Fingerprint of IPub2                                         |
|             From | Generate IPriv1, then IPub1 in this field                    | Generate CPriv2, then CPub2 for this field                   | Generate IPriv2, then IPub2 for this field                  | Generate CPriv3, then CPub3 in this field                    |
|          Message | Signed with IPriv1, encrypted with Cipher1 from CPub1 and IPub1 ECDH | Signed with CPriv2, encrypted with Cipher2 from CPriv2 and IPub1 | Signed with IPriv2, encrypted Cipher3 from IPriv2 and CPub2 | Signed with CPRiv3, encrypted with Cipher4 from IPub2 and CPriv3 |
|          Expires |                                                              | List of fingerprints of public keys in From fields seen at time of sending | “”                                                          | “”                                                           |
| Packet Signature | Whole message data signed by IPriv1                          | Whole message data signed by CPRiv2                          | Whole message data signed by IPriv2                         | Whole message data signed by CPriv3                          |

In the case of sending multiple messages before reply, the party generates new ciphers for each one, and cipher is created from known public key of the other party.

All keys generated for such messages must be kept for a time. When a reply is composed, all known received key fingerprints are appended to end of message to indicate they can be purged, in addition to the fingerprint in the To field.

## Message Binary Format

| Field        | Size                      | Description                                                  |
| ------------ | ------------------------- | ------------------------------------------------------------ |
| To           | 8 bytes                   | Fingerprint of public key of recipient used in with ECDH for cipher |
| From         | 32 bytes                  | Public key of sender used with ECDH for cipher               |
| MessageNonce | 12 bytes                  | Cryptographically random nonce for message encryption        |
| MessageSize  | 4 bytes                   | Size of message (up to 4Gb)                                  |
| Message      | variable, per MessageSize | Message is signed, signature appended, then encrypted with cipher |
| ExpireCount  | 2 bytes                   | Number of expired public keys seen prior to dispatch of this message |
| Expired      | 8 bytes                   | Fingerprint of expired public keys of recipient that have been seen |
| …            | repeats per ExpireCount   |                                                              |
| Signature    | 64 bytes                  | Signature over entire message data (all previous fields) to prevent tampering |

