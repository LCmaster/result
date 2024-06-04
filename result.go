package result

import (
	"errors"
	"fmt"
)

type ResultHolder[T any] interface {
	IsOk() bool
	IsError() bool
	Get() (T, bool)
	GetError() error
	OrElse(T) T
	OrElseGet(func() T) T
	IfOk(func(T))
	IfError(func(error))
}

type Result[T any] struct {
	value *T
	err   error
	ResultHolder[T]
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}
func (r Result[T]) IsError() bool {
	return !r.IsOk()
}
func (r Result[T]) Get() (T, bool) {
	return *r.value, r.IsOk()
}
func (r Result[T]) GetError() error {
	return r.err
}
func (r Result[T]) OrElse(other T) T {
	if r.IsOk() {
		return *r.value
	}

	return other
}
func (r Result[T]) OrElseGet(supplier func() T) T {
	if r.IsOk() {
		return *r.value
	}

	return supplier()
}
func (r Result[T]) IfOk(consumer func(T)) {
	if r.IsOk() {
		consumer(*r.value)
	}
}
func (r Result[T]) IfError(consumer func(error)) {
	if r.IsError() {
		consumer(r.err)
	}
}

func (r Result[T]) String() string {
	if value, ok := r.Get(); ok {
		return fmt.Sprintf("%v", value)
	}

	return r.err.Error()
}

func From[T any](value T, err error) Result[T] {
	return Result[T]{value: &value, err: err}
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: &value, err: nil}
}

func Error[T any](err string) Result[T] {
	return Result[T]{value: nil, err: errors.New(err)}
}

func Try(e error, success func(), catch func(error)) {
	if e == nil {
		success()
	} else {
		catch(e)
	}
}
