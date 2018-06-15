package errors

import (
	"fmt"
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
type Handler struct{}

// Catch catches errors in the function it was called in it. The default
// behavior is to catch all errors except runtime.Error, and it goes like this:
//
//	defer h.Catch(&err)
//
// You can pass to it only specific errors if you want to catch only those. For example:
//
//	defer h.Catch(&err, io.EOF, &time.ParseError{})
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
	panic(rerr)
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

// Must panics if error occurred.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Assert panics if the assertion is false.
func Assert(cond bool, format string, v ...interface{}) {
	if !cond {
		panic(AssertError{
			fmt.Sprintf(format, v...),
		})
	}
}

// indirect returns the item at the end of indirection.
func indirect(t reflect.Type) reflect.Type {
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
	}
	return t
}
