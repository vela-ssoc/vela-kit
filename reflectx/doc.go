// Package xreflect simplifies data passing to and from gopher-lua.
// (https://github.com/yuin/gopher-lua).
//
// # Go to Lua conversions
//
// See documentation of ToLValue function.
//
// # Lua to Go conversions
//
// Lua types are automatically converted to match the output Go type (e.g.
// setting a struct field from Lua).
//
// lua.LNil can be converted to any channel, func, interface, map, pointer,
// slice, unsafepointer, or uintptr value.
//
// lua.LBool values are converted to bool.
//
// lua.LNumber values are converted to float64.
//
// lua.LString values are converted to string.
//
// lua.LChannel values are converted to lua.LChannel.
//
// *lua.LTable values can be converted to an array, slice, map, struct, or
// struct pointer. If the table is being assigned with no type information (i.e.
// to an interface{}), the converted value will have the type
// map[interface{}]interface{}.
//
// The Data field of *lua.LUserData values are converted rather than the
// *lua.LUserData value itself.
//
// *lua.LState values are converted to *lua.LState.
//
// *lua.LFunction values are converted to Go functions. If the function is
// being assigned with no type information (i.e. to a interface{}), the function
// will have the signature func(...interface{}) []interface{}. The arguments
// and return values will be converted using the standard xreflect conversion rules.
//
// # Thread safety
//
// This package accesses and modifies the Lua state's registry. This happens
// when functions like ToLValue are called, and potentially when xreflect-created values
// are used. It is your responsibility to ensure that concurrent access of the
// state's registry does not happen.
package reflectx // import "layeh.com/gopher-xreflect"
