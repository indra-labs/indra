# splicer

Splicer is a simple library that makes use of interfaces and dynamic arrays (slices) to simplify the assembly of messages containing several fields in a
specified order

To encode this to a binary format, and this format includes the
offsets of each element so that rather than decoding in a batch, decoding can be
selective, an important feature for a low latency message processing handler.

The structure consists of 3 values at the top, and the structure of the remainder is understood by the code.

1. a sentinel, called the "magic bytes" - a common structured binary document convention, also using the conventional 32 bit value with 4 bytes, that usually will also be human readable ASCII strings.
2. The number of fields in the message, which enables the selective decoding of elements of the message.
3. The offsets of the fields in the message come next, as a series of 32 bit values.

The remainder of the message then contains the segments specified via the offsets where each numbered element of the field can be decoded from, available via an accessor method on the container type.

This library was inspired by the Flatbuffers binary encoder, which similarly structures the data so that the fields can be independently decoded. It does not progressively encode, there is no reason for this since the entire message must be sent out.

It differs from Flatbuffers in that it does not require any kind of parser, and aside from the segment information, the messages have no other metadata, the receiver must have the defined structure that matches the magic, and without this key matching none of the data can be accessed.

The reason for this type of progressive decoding is twofold:

1. It eliminates the need to allocate any memory other than to decode the fields that are accessed, which populate a new variable corresponding to the binary form, and if the received message needs to be persisted or forwarded, it can be thus processed without any extra work.
2. Because it is for a message relaying network and a strict policy of fail fast means that when messages are not valid, or not authorised, or other reasons that can be quickly determined, the time spent decoding is limited to only decoding the fields needed to pass through to further processing if needed, or able to be aborted and the buffer and the processing time given over to other messages, making the latency as low as possible.

### Varints

While it would be possible to include this type, as can be seen looking through [../../docs/formats.md](../../docs/formats.md) there just isn't any real need for it, it would scarcely save any space.
