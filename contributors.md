### contributor guidelines

# Indranet Project Contributor Guidelines

## [Programming] Languages

The less, the merrier.

We prefer Go, but for initial reference implementations or experiments where
needed an FFI via CGO can be used. This enables at least C, C++ and Rust to be
used by Go.

It is preferred to not do this if at all possible, because these functions
become a black hole of non information inside the profiler. Concurrency
bottlenecks can sometimes be hard to find, even if the code appears to be
correct and well constructed.

If at all possible, outside of this, restrict foreign language usage
to `sh`/`bash` scripts, `html`/`css`/`javascript`, which may be needed for all
kinds of reasons, but we also want to stress, preferably not a web interface,
just for the security of users.

GUI libraries are available for Go, but if it really doesn't cut it, design the
foreign language GUI part as a simple set of controls for an API to the
component you are adding/building.

For wire encoding, outside of the constraints of the compact and canonical
formats used in our onions and other messages, if you intend that this interface
might connect with any number of different other peers, then use `gRPC/Protobuf`
and ideally we will centralise the protocol files and their generated outputs in
a similar way as recommended below in the section about Interfaces.

## Comments

No code should end up on the `master` branch without having a properly formed go
documentation comment on the exported methods.

This can be excepted in the case of interfaces, where the interface and its
methods must bear substantial documentation and the implementations should have
it where important things about it need to be known by users about how it
differs from the version on the interface. And in that case, there will be
questions, and either the interface comments amended or the implementation is
made conformant.

## Whitespace, Margins and Function Signatures

In general, you can put empty lines in as much as you like, but it is preferred
if you adhere to a few rules:

- ### 4 spaces = 1 tab indent

  > 8 is too big. Probably even it's from typewriters and even old TTYs usually
  had 120 width.

- ### 80 is the width, and the number of the width is 80, not 81, not 95.

  > If your function has turned into cancer, and the indents leave you no space
  > for the necessary names, make a new function and wrap this up in it. Thank
  us
  > later when you reuse it.
  >
  > The functional, neurology reason for this is the same as the page width and
  column width used in novels and newspapers. Long lines are visually harder to
  return to the left column on the correct line. When the line is no more than
  7-10 words long, it is easy, as it is not quite yet fully in your peripheral
  vision at this point.

- ### Function Signatures also the rule of 80, but try to avoid breaking except on commas in parameter and return lists, and newline if the function breaks, to keep it visually separate.

  > Like you will see repeatedly throughout these recommendations, these themes
  repeat and they repeat because the same principles are at play. **The less
  time it takes to recognise something, the more often you find incorrect things
  and often, squash bugs.**

- ### When methods are very short, group them and one-line the single statement in it, if it fits in 80 columns

  > Unfortunately, this isn't consistently available in the `gofmt` code layout
  rules for `if` or `for` blocks but there is many times functions in
  implementations which are a little orthogonal to the type. This is also why
  interfaces should be small. If you find yourself making a lot of no-ops that
  make nice one-liners by `gofmt` maybe better to break it into two and embed
  the base into the less common but bigger interface.
  >
  > *But why 80 characters wide*, do you say? *We have fancy 5k pixel wide
  > displays*.
  >
  > Yes, maybe you do, but more often it is most useful to see code side by side
  > than vertically, for fast pattern recognition.
  >
  > Or just to make it easy to read the documentation comments the last person
  left
  > for you to enjoy, and use it for the first time in a piece of code.
  >
  > If 80 columns is stifling for you, you are using excessively long and
  precise
  > names, and writing functions that should be broken into components.

- ### Composition > Inheritance - Decompose interfaces, and consolidate repeating fields in structs.

  > Interfaces and anonymous types allow Go to enable extension of existing code
  without modification, without incurring the additional complexity of both
  translating such nebulous and excessively complex designs that lead to long
  compilation times or poor execution, or forcing us to specify the anonymous
  field name (the type), and to the end user programmer, the interface tells
  all, the details are not important if they are not impediments.

## Interface Usage

At Indranet, we love interfaces. Network interfaces, the avatar of a user on a
peer to peer network (identity), graphical user interfaces, control interfaces
and the interface between us animals and these artefacts of our cleverness, and
even the interfaces between us animals and between peers of a species.

Go lets you use them, and is also kind enough to not subject you to massive
memory use, long compilation times, redundant .h files, and the list is endless
for us systems programmers. It stifles flow, and flow is important to creativity
and focus.

As a part of our project we need to help recommend how, and give the reasons
why, that if a contributor puts something into a PR that should have an
interface associated with it, because we think that unlike humans, partnerships
between programmers are not marriages and interfaces give future developers, and
us, and you, the ability to add a new interface and then the magic is you change
a few references for the implementation and **presto chango**, new
implementation
slotted in.

- ### Create an interface for every type you will use a lot in other code early, if not right at the beginning.

  > This can save a lot of time caught in rabbitholes, it acts like guiderails.

- ### Place the methods of the interface into a single file separate from th  methods and functions used in the first implementation, ideally, in one place  for the entire repository.

  > By putting each in its own file, it can be seen from the folder how many are
  > there, and what they are.

- ### Do write functions that operate purely on the interface, and put them in a subfolder of the interfaces under a package name form of the interface name.

  > These functions cannot be written as methods because you can't declare
  > methods on an interface type, even within a package.

- ### Amend the interface when it appears that things initially thought internal are internal and need to be used as interfaces.

  > The signature of the need for an interface in Go is recurrance of parameters
  > of functions, though the return values also. When it is a conflict between one
  > that returns no error, and one that returns an error, change all to the one
  > with the error as well as the return value. **These can get awkward, but
  > usually the worst of it is avoided by putting a string of error handling
  > invocations that recur in a package within a helper function.**

- ### Use meaningful variable names in the interface prototype to evoke the meaning of the parameter.

  > Get into the habit of avoiding anonymous return values. If functions are
  > short, and they ideally should be less than a screenful or about 60 lines
  > long, then the other benefit is you can use naked returns without confusing
  > people - because the return values have meaningful names, there at the top.

- ### Comment them well to express the contract where it is ambiguous or uncertain what may be intended in the more marginal optional elements.

  > This also lets you get away with not documenting the implementations'
  > methods, since tracing the interface shows you its documentation. Goland
  > facilitates this, if your IDE doesn't, it's really worth the 10 euro/month
  > license.

- ### There can be uses for unexported interfaces within packages, but more than likely the package is getting overgrown with types at that point, in parallel.

  > To be perfectly honest, it is paradoxical to talk about interfaces (outside
  > surface between two things) and encapsulated.
  >
  > *YMMV but most of the time it will be beneficial to separate out some of it
  > and make it more friendly to the human short term working memory's
  > limitations.*

Having a clear view of the interfaces in the system helps immensely in getting
that top down architect view that act as guard rails from elaborating messy or
inconsistent methods in the interface bearing type.

Interfaces also help you contain features, if you think of them as being the
prototypes of the layers of your system. If you don't run away with yourself in
implementing things and first focus on interfaces you get the job done sooner
too. And those other things can become their own thing later, when the MVP is
ready.

Interfaces are unpopular outside of Java and Go and like CSP and channels, big
elements of the rationale for using Go and preferring Go.

They do most of the same things that generics, objects and templates achieve,
but are far simpler and are simply a form of indirection like pointers, with a
type to indicate the base set of methods are for a given type in a package.

As such they still need switches but a well defined and sufficient interface
essentially lets you avoid a lot of repetition as well as to make code that uses
the interfaces drop in, in place or alongside other code that also uses it, with
no changes, even if the concrete types involved are different to what was used
in the first instance. Go uses this extensively in the `io` and `net` libraries.

Moving the PC of a CPU's processing thread from one place to another is the
least overhead operation a CPU does, when it gets a cache hit or the cached data
is not far upstream towards the main memory. It is far cheaper than the complex
switching logic that goes on behind the scenes in dynamic types.

Thus it is simpler, and less expressive, but it does all the jobs you need it to
do, as long as you architect with it in mind. 99% of cases the gripes are just
*wrong* and lazy.

## Epilogue

Many luke-warm, wishy washy, wide eyed, grinning mouth breathers may find our
opinions too firm, well, that's a shame that you don't like the engineer
mindset, cos it would be needed to work on Indranet. Security and Privacy are
not just brand names. They are the product of a mindset that is required to
efficiently implement the designs and effectively yield the services users
expect.

These guidelines are quite strong, almost Laws, and they will be enforced
because of the greater efficiencies they enable. There is good science behind
many of the principles in the understanding of the function of our brains in
neurology, matters like visual pattern recognition, which can save a lot of
time, the limitations of short term working memory and it's concurrent moving
parts (for most, about 7 items), and counting - we can only instantly
distinguish 1, 2 and 3, and have to chunk down to count more.

And outside of the human/code/machine interface, the facts of how computer
hardware functions, and the things that it does poorly, and the considerations
of the amount of latency created as assets fill up the space, and the complexity
of data and it's semantic graph and complexity and how that tends to translate
to slower, memory chewing, GC pausing pain.

As regards to etiquette, we aren't gonna codify that, because it's simple.
Critique the work, not the person, and don't try to partition us as we have
protocols to detect and eject such invasions. We are veterans and avid students
of the arts of deception and persuasion, and know a fallacy when we hear one.
Don't waste your time.

> # “If you don't believe me or don't get it, I don't have time to try to convince you, sorry.”
>
> satoshi's answer to bytemaster about block times

Anyone who knows who `bytemaster` is and the rackets he ran since that time will
get it.

We don't apologise for requiring the best, most reasonable policies on our code
and operational communications, because if they don't work, our enemies will
profit.

We aren't recruiting contributors for the same reasons as we would want to apply
to the users, whose use is the pure metric of success of this project.

We want you to use these guidelines to help us bring the most secure, performant
and widely useful feature of privacy to the internet, in the shortest possible
time, and that means we aren't wasting our time any further about the rationale
for our policies.

You are most free to have a go at presenting a point that actually we get within
a 5 minute interaction, and if you are right and we were legalistic pigs, we
will come back to you, and the amendment will be published.

Intellectual pursuits are no less brutal than team contact sports. We do this
because better to argue now than to be defeated later on.

Too many rules and creativity is squashed. Not enough rules and nothing of
substance is ever done. Just as the most interesting things in the universe are
the interfaces between things, so too the rules are an interface between keeping
order and efficiency, and enabling creativity and flexibility. And thus also, it
helps a lot to put them up front and give them in a friendly tone.