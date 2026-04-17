// Package result provides a generic Result[T] type that represents either a
// successful value of type T or an error. It is inspired by the Result<T, E>
// type found in Rust and the Optional/Either monads in functional languages,
// offering a more explicit, type-safe alternative to the idiomatic Go
// (value, error) return pair.
//
// # Quick start
//
//	import "github.com/LCmaster/result"
//
//	// Wrap an existing (value, error) pair:
//	r := result.From(strconv.Atoi("42"))
//
//	// Construct explicitly:
//	ok  := result.Ok(42)
//	err := result.Error[int](errors.New("not found"))
//
//	// Consume safely:
//	if v, ok := r.Get(); ok {
//	    fmt.Println("parsed:", v)
//	}
//	fmt.Println(r.OrElse(0))
//
// # Using ResultHolder as an interface
//
// [ResultHolder] is the interface implemented by [Result]. Depend on it in
// function signatures when you want to accept any result-like value without
// coupling to the concrete type:
//
//	func process(r result.ResultHolder[string]) {
//	    r.IfOk(func(s string) { fmt.Println(s) })
//	}
package result

import (
	"fmt"
)

// ResultHolder is the interface satisfied by [Result]. It exposes the full
// read API without exposing implementation details, making it suitable for use
// in function parameters, return types, and mocks in tests.
type ResultHolder[T any] interface {
	// IsOk reports whether the result holds a value (no error occurred).
	IsOk() bool

	// IsError reports whether the result holds an error (no value).
	IsError() bool

	// Get returns the held value and true when the result is ok.
	// When the result is an error, it returns the zero value of T and false.
	// It never panics.
	Get() (T, bool)

	// Error returns the underlying error, or nil when the result is ok.
	Error() error

	// OrElse returns the held value when ok, or other when an error.
	OrElse(other T) T

	// OrElseGet returns the held value when ok, or the value produced by
	// supplier when an error. The supplier is only called when needed,
	// making it preferable over OrElse when computing the fallback is costly.
	OrElseGet(supplier func() T) T

	// IfOk calls consumer with the held value only when the result is ok.
	// It is a no-op when the result is an error.
	IfOk(consumer func(T))

	// IfError calls consumer with the error only when the result is an error.
	// It is a no-op when the result is ok.
	IfError(consumer func(error))
}

// Result holds either a value of type T or an error, but never both.
// Use the constructor functions [Ok], [Error], [From], or [FromPtr] to create
// a Result. The zero value of Result is not a valid state.
//
// Result[T] implements [ResultHolder[T]] and [fmt.Stringer].
type Result[T any] struct {
	value *T
	err   error
}

// IsOk reports whether the result holds a value (i.e., no error occurred).
//
//	r := result.Ok(1)
//	r.IsOk() // true
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsError reports whether the result holds an error (i.e., no value).
//
//	r := result.Error[int](errors.New("oops"))
//	r.IsError() // true
func (r Result[T]) IsError() bool {
	return !r.IsOk()
}

// Get returns the held value and true when the result is ok.
// When the result holds an error it returns the zero value of T and false.
// Get never panics.
//
//	if v, ok := r.Get(); ok {
//	    fmt.Println("value:", v)
//	} else {
//	    fmt.Println("error:", r.Error())
//	}
func (r Result[T]) Get() (T, bool) {
	if r.value == nil {
		var zero T
		return zero, false
	}
	return *r.value, true
}

// Error returns the underlying error, or nil when the result is ok.
//
//	r := result.Error[int](io.EOF)
//	errors.Is(r.Error(), io.EOF) // true
func (r Result[T]) Error() error {
	return r.err
}

// OrElse returns the held value when the result is ok, or other otherwise.
// Use [OrElseGet] instead when computing the fallback value is expensive.
//
//	n := result.Error[int](err).OrElse(-1) // n == -1
func (r Result[T]) OrElse(other T) T {
	if r.IsOk() {
		return *r.value
	}

	return other
}

// OrElseGet returns the held value when the result is ok, or the value
// produced by calling supplier otherwise. The supplier is only called when
// the result is an error, so it is safe to use for expensive computations.
//
//	n := r.OrElseGet(func() int { return expensiveDefault() })
func (r Result[T]) OrElseGet(supplier func() T) T {
	if r.IsOk() {
		return *r.value
	}

	return supplier()
}

// IfOk calls consumer with the held value if the result is ok.
// It is a no-op when the result holds an error.
//
//	r.IfOk(func(v int) { fmt.Println("got:", v) })
func (r Result[T]) IfOk(consumer func(T)) {
	if r.IsOk() {
		consumer(*r.value)
	}
}

// IfError calls consumer with the error if the result is an error.
// It is a no-op when the result is ok.
//
//	r.IfError(func(e error) { log.Println("error:", e) })
func (r Result[T]) IfError(consumer func(error)) {
	if r.IsError() {
		consumer(r.err)
	}
}

// String implements [fmt.Stringer]. It returns the string representation of
// the held value when ok, or the error message when an error.
func (r Result[T]) String() string {
	if value, ok := r.Get(); ok {
		return fmt.Sprintf("%v", value)
	}

	return r.err.Error()
}

// From wraps a standard Go (value, error) return pair into a Result.
// If err is non-nil the value is discarded and an error Result is returned;
// otherwise an ok Result holding value is returned.
//
// This is the most convenient constructor when adapting existing Go functions:
//
//	r := result.From(strconv.Atoi(input))
//	r := result.From(os.Open(path))
func From[T any](value T, err error) Result[T] {
	if err != nil {
		return Error[T](err)
	}
	return Ok[T](value)
}

// FromPtr wraps a (*T, error) pair — common in database or repository
// functions — into a Result. When ptr is nil an error Result is returned:
// the provided err is used if non-nil, otherwise a synthetic error is created.
//
//	row, dbErr := db.QueryRow(...)
//	r := result.FromPtr(row, dbErr)
func FromPtr[T any](ptr *T, err error) Result[T] {
	if ptr == nil {
		if err == nil {
			err = fmt.Errorf("FromPtr: nil pointer with no error")
		}
		return Error[T](err)
	}
	return Ok[T](*ptr)
}

// Ok constructs a successful Result holding value.
//
//	r := result.Ok(42)
//	r.IsOk() // true
func Ok[T any](value T) Result[T] {
	return Result[T]{value: &value, err: nil}
}

// Error constructs an error Result holding err.
// The type parameter T must be specified explicitly when it cannot be inferred:
//
//	r := result.Error[int](errors.New("not found"))
//	r.IsError() // true
func Error[T any](err error) Result[T] {
	return Result[T]{value: nil, err: err}
}

// Try is a control-flow helper that runs success when e is nil, or iterates
// through the optional catch handlers when e is non-nil.
//
// Each catch handler receives the error and returns a bool indicating whether
// it handled the error. Iteration stops at the first handler that returns true.
// If no handler returns true the error is silently ignored.
//
//	result.Try(err,
//	    func() { fmt.Println("success") },
//	    func(e error) bool {
//	        if errors.Is(e, ErrNotFound) {
//	            fmt.Println("not found")
//	            return true // handled
//	        }
//	        return false // pass to next handler
//	    },
//	)
func Try(e error, success func(), catch ...func(error) bool) {
	if e == nil {
		success()
	} else {
		for _, errCatcher := range catch {
			if errCatcher(e) {
				return
			}
		}
	}
}
