# Holo

Holo is an algorithm for storing large amounts of text using prime numbers
and powers as graph vectors.

Conventional programming for boolean value storage comes in two flavours,
the C style 1 byte 1 value, and the older style assembler bit flag filter.

The efficiency of extracting the value is lower for a bitfield, because you
have to copy the byte and then perform an AND using the true value for the flag.

The byte, like as used in Go, is just a comparison to the true value 1 and
only uses one instruction, where the bitfield must first mask and then
compare to zero.

Except for embedded applications, there usually is not much use for the
trivial memory savings for a relatively small number of values a program can
need that store a true/false value.

## Text Compression

It is well known that text content has a very high level of redundancy in
its structure, and of course most compression algorithms do a pretty good
job with them, but the encoding systems are not built to allow you to
traverse the document one symbol at a time, they just spit the reconstructed
version at whatever file you point them at.

Holo is designed *only* for text, performs a greedy, smaller up to larger
byte length scanning of a chunk of data and generates a dictionary of 256
bit holo vectors associated with each of the dictionary entries, and it
expects text like structure with whitespace, symbols, numbers and letters as
the 4 categories of byte types. Spaces after a symbol are stored also in the
filter, as are paragraph boundaries.

## Prime Numbers as Compact Bitfields

The specific property required for a bit field is that altering one bit in a
field has no impact on any other bit. Thus, the traditional AND masking and
zero comparison for boolean values satisfies this, as flipping a bit does
not alter the other bits. Obviously byte-as-boolean can be operated on with
a single comparison and no bitwise masking, faster, but more memory used.

If your data involves thousands or millions of binary values, the
conventional idea is you are gonna have to use up a few hundred kilobytes
and a big map of symbols to bit positions.

But there is a better way.

The strange and interesting thing about prime numbers is that they can be
composited together in a single number, and all primes except 1 can be
multiplied into your number, and performing a modulo division (the remainder)
with a given prime number yields no remainder if it is present, or a nonzero
remainder if it is not.

There is millions of prime numbers, and way way more than you could ever
need in a 256 bit number. Even a 64 bit number can give you a lot of prime
values all squished together in one little double long word.

## More than Binary

The presence or absence of a particular prime factor can be more than just a
true/false value. It is also possible to exponentiate these values, and they
can then, in addition to the simple boolean present/not present option,
there can be also 1, 2, and more values at each prime number, essentially
enabling also ternary, quaternary, etc values, limited by the precision of
the numbers and the size of the prime number compared to the size of the bit
field.

So, in addition to having a present/missing state for a ridiculously large
number of distinct binary values, they can be multi-state.

It seems incredible but it is possible to thereby store enormous amounts of
data in these fields.

## The Inspiration

Pondering upon the architecture of Transformer style text generators like
OpenAI's GPT, I had the idea that somehow one could store a document as a
dictionary from its distinct and longest repeating sequences, and the lists
of positions of the repeating symbol elements in the text that can be used
to seek forward and backward in the original text from which the holo matrix
is derived, by following the document symbol positions referred to in the
256 bit number fields storing each symbol's set of next symbols.

The prime fields can be done over other sizes than 256 bits, but 256 bits is
necessary to allow enough headroom above the prime numbers to prevent an
overflow and loss of data, for compressing large volumes of text.

Most texts contain less than 10,000 distinct words and about half of them
are repeated several times in a line. So if the average word length is about
5 characters, the symbol part of the dictionary is only going to be a few
hundred kilobytes and up to a few megabytes for a really broad and large
text selection or one containing multiple languages.

Because Holo is aimed at very large text compression, it assumes that all
words are space or symbol separated, and uses some of the primes to store
pre- and post- whitespace, tabs or newlines in the word "symbol" dictionary
prime field.

A document has a first symbol, and from this first symbol one can
reconstitute the original document, with those structuring assumptions based
on text.

The prime field for each symbol, after the first 6 primes 2, 3, 5, 7, 11,
13, which represent space, tab, linefeed before, and after, are a coordinate
space based on the sequence of prime numbers mapped to normal ordinals, each 
other symbol in the dictionary represented by the first row, and the 
sequence number of these primes corresponds to each symbol in the dictionary.

Then there is more rows at each multiple of the dictionary's entry size, 
these subsequent rows representing the sequence of outbound links to 
subsequent symbols in the document. As the Holo iterator walks the document, 
it must keep a track of the prime sequence number along the path to enable 
reverse. The symbol found at a given document symbols position should be 
cached every 100th symbol or so, and after the first walkthrough fast random 
seeking becomes possible.

An interesting thing about Holo also is that it can be edited, since all 
that is required is to change what symbol a given document position contains,
inserting is non-destructive though one must add the branch to the inserted 
new symbol, and it must then point back to the previous subsequent symbol.

It is also possible to add or remove symbols, though this requires a full 
iteration.

## Planned Uses

Holo, in theory, should be able to compress large bodies of text massively, 
while retaining reasonably efficient seek times. So, for example, an entire 
program application source code could be stored in a Holo Grid.

The interrelations between symbols, the raw frequency information of the 
sequences, alone, provides a minimal statistical probability for one word to 
follow another, this is simple to discover and enumerate, although it 
requires a full walk of the path.

Grammar information is indirectly stored also, because the majority of 
sequences of symbols point to a next symbol that is of a given part of 
speech, like for example how the definite article "The", is usually followed 
by a noun, or a composite noun with chains of nouns adjectives and nouns.

Alone, by the weightings of the links from one symbol to another in the 
document sequence, one can greatly thin out the options for a valid next 
word given a current word.

The grammar, the patterns of different parts of speech found in a language, 
form repeating sequences of symbols, after tokenising the text, can be 
searched in the same way as the sequences of bytes are used to gather the 
words, whitespace, symbols and numbers, with multiple symbols repeating 
sequence.

The compression of multiple symbols both saves on space and provides a graph 
between the words that not only shows the probability of any given 
subsequent symbol (word), but can show that a series of longer sequences 
that have partial matches with things like "The X of Y", this prime field is 
thus storing enough information that it can take an input, generate its own 
independent dictionary and path vectors, and then show you a list of the 
words that it has observed in a repeating partial match fragment, and by 
inference, determine that the symbols that differ between the two otherwise 
similar patterns are the same part of speech.
