// Package splice provides a series of interfaces and abstract methods on them
// for implementing a simple cursor based read/write interface for slices of
// interfaces that enable composition, editing and decoding of composite data
// types.
//
// Note that a key feature of this design is the elimination of the rigid
// structure of `struct` declarations in order to reduce the excess boilerplate
// to handle these messages composed of strings of elements. Instead,
// implementations that provide concrete types (both singular and structured)
// will "declare" these structures from a function that returns the structure
// containing empty types, in much the same way as a struct does, except for the
// additional element of indirection for the interface.
//
// By using slices of a concrete interface type (not the empty interface) the
// possibility of quickly creating new message types becomes easy and does not
// require altering this complex of type abstractions provided in this package.
package splice
