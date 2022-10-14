# simplebuffer

For simple applications it makes no sense to have complicated things added to
them. This library aims to solve two problems in one: A binary encoding with
a decoder that avoids copying and not using reflect but not making new data
types difficult to add or new messages to string together. Something fast but
simple and easy to extend and recompose.
 
Using a slice of a specific serializer type, with an interface that does the
usual things required of a serializer, mainly, to produce a serialized
representation of its contents, can be strung together in an arbitrary order
and do not run until you generate the message container, and when received do
not incur processing cost until the value is actually accessed, which can
help ensure failure conditions are more rapidly responded to - only as many
fields are needed to determine such a condition are required to be decoded
in order to respond appropriately, whereas a standard serialization codec
will decode everything anyway before you even see it.

This library is inspired by Flatbuffers, but its terrible support of Go and
the use of complicated tools and syntaxes and generators, when I already
understand interfaces and slices and before there is a version 1 it is easy
to just ad-hoc write new types. The design is such that also when you use it,
each interface implementation is a unit of compilation, and not any of the
implementations present that you haven't imported.
 
In most cases, there is already a fast binary encoding scheme for almost any
given data type implemented by servers already existing, usually fast, this
allows also the arbitrary change of the content of messages for specific
purposes without a performance hit from reflection. It
is up to the user of the library to take care of message versioning and
the simplest way would be to centralise the definitions of the magics.
The encoders can be made zero-copy at the option of the implementor, the
performance cost of copying is still less than reflection anyway, but the
option is open for the future.  
 
This is not really so much of a library as a pattern of construction with
modular importable boilerplate. The goal is maximum performance especially
latency in tasks that may not require complete decoding to determine a
response. 

Go is a language that encourages creative repurposing of its syntax to many
other things than what was intended, and the use of closures, first class
functions and parser/generators stringing these calls enables metaprogramming
, yet without imposing complicated new syntax or feeling excessively clunky
 (once you get over the `func` :) )