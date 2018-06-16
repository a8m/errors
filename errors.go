package errors

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
)

// AssertError returns by the Expect method if the given assertion is false.
type AssertError struct {
	msg string
}

func (a AssertError) Error() string { return a.msg }

var assertType = reflect.TypeOf(AssertError{})

// Handler is the type you embed in your struct in order to give it the "fancy" errors handling
// flow control.
type Handler struct {
	// Panic is called when an error that was not expected has been called.
	// If not defined, the standard "panic" will be called.
	Panic func(error)
}

// Catch catches errors in the function it was called in it.
// If no types are given, it catches all errors except runtime.Error.
// Otherwise it catches only the defined error types.
// Usage:
//
//      defer h.Catch(&err)
//
// Or:
//
//	    defer h.Catch(&err, io.EOF, &time.ParseError{})
//
func (h *Handler) Catch(err *error, types ...error) {
	r := recover()
	// no error occurred.
	if r == nil {
		return
	}
	// what are you throwing?
	rerr, ok := r.(error)
	if !ok {
		panic(rerr)
	}
	// if no types were defined, catch all except runtime errors.
	if len(types) == 0 {
		// don't catch runtime errors.
		if _, ok := rerr.(runtime.Error); ok {
			panic(rerr)
		}
		*err = rerr
		return
	}
	// postpone the usage of reflection to the end.
	typ := indirect(reflect.TypeOf(rerr))
	// if is an AssertError type, return it
	if typ == assertType {
		*err = rerr
		return
	}
	// if the error is one of the defined types, return it
	for _, t := range types {
		if indirect(reflect.TypeOf(t)) == typ {
			*err = rerr
			return
		}
	}
	if h.Panic == nil {
		panic(rerr)
	}
	h.Panic(rerr)
}

// Must panics if error occurred. Should be caught by Catch.
func (h *Handler) Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Assert panics with a custom error. For example:
//
//	type Parser struct {
//		errors.Handler
//	}
//
//	p := new(Parser)
//
//	p.Assert(len(input) > 0, &ParseError{msg: "empty input"})
//
func (h *Handler) Assert(cond bool, err error) {
	if cond {
		return
	}
	panic(err)
}

// Assertf panics with AssertError. For example:
//
//	type Parser struct {
//		errors.Handler
//	}
//
//	p := new(Parser)
//
//	p.Assert(len(input) > 0, "empty input")
//
func (h *Handler) Assertf(cond bool, format string, v ...interface{}) {
	if cond {
		return
	}
	msg := fmt.Sprintf(format, v...)
	panic(AssertError{msg})
}

// defaultHandler is used for package level functions.
// It changes the behavior when meeting an error that was not expected to
// log.Fatal instead of panic.
var defaultHandler = Handler{
	Panic: func(err error) { log.Fatal(err) },
}

// Must panics if error occurred.
func Must(err error) {
	defaultHandler.Must(err)
}

// Assertf panics with AssertError if the assertion is false.
func Assertf(cond bool, format string, v ...interface{}) {
	defaultHandler.Assertf(cond, format, v...)
}

// Assert panics with a custom error if the assertion is false.
func Assert(cond bool, err error) {
	defaultHandler.Assert(cond, err)
}

// Catch catches errors in the function it was called in it.
// If no types are given, it catches all errors except runtime.Error.
// Otherwise it catches only the defined error types.
func Catch(err *error, types ...error) {
	defaultHandler.Catch(err, types...)
}

// indirect returns the item at the end of indirection.
func indirect(t reflect.Type) reflect.Type {
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
	}
	return t
}
