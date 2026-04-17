package result

import (
	"errors"
	"fmt"
	"testing"
)

var errSentinel = errors.New("sentinel error")

// ── Constructors ────────────────────────────────────────────────────────────

func TestOk(t *testing.T) {
	r := Ok(42)
	if !r.IsOk() {
		t.Fatal("Ok() result should be ok")
	}
	if r.IsError() {
		t.Fatal("Ok() result should not be an error")
	}
	if r.Error() != nil {
		t.Fatalf("Ok() result error should be nil, got %v", r.Error())
	}
	v, ok := r.Get()
	if !ok {
		t.Fatal("Get() on Ok() result should return true")
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestError(t *testing.T) {
	r := Error[int](errSentinel)
	if r.IsOk() {
		t.Fatal("Error() result should not be ok")
	}
	if !r.IsError() {
		t.Fatal("Error() result should be an error")
	}
	if !errors.Is(r.Error(), errSentinel) {
		t.Fatalf("expected sentinel error, got %v", r.Error())
	}
}

func TestFrom_ok(t *testing.T) {
	r := From(99, nil)
	if !r.IsOk() {
		t.Fatal("From(v, nil) should be ok")
	}
	v, ok := r.Get()
	if !ok || v != 99 {
		t.Fatalf("expected (99, true), got (%d, %v)", v, ok)
	}
}

func TestFrom_error(t *testing.T) {
	r := From(0, errSentinel)
	if r.IsOk() {
		t.Fatal("From(v, err) should not be ok")
	}
	if !errors.Is(r.Error(), errSentinel) {
		t.Fatalf("expected sentinel error, got %v", r.Error())
	}
}

func TestFrom_doesNotPreserveValueWhenErrorSet(t *testing.T) {
	// Previously both value and err were stored simultaneously (invalid state).
	// Now, value must be discarded when err is non-nil.
	r := From(123, errSentinel)
	v, ok := r.Get()
	if ok {
		t.Fatal("Get() should return false when From has an error")
	}
	var zero int
	if v != zero {
		t.Fatalf("Get() should return zero value on error result, got %d", v)
	}
}

func TestFromPtr_ok(t *testing.T) {
	n := 7
	r := FromPtr(&n, nil)
	if !r.IsOk() {
		t.Fatal("FromPtr(&v, nil) should be ok")
	}
	v, ok := r.Get()
	if !ok || v != 7 {
		t.Fatalf("expected (7, true), got (%d, %v)", v, ok)
	}
}

func TestFromPtr_nilPtr_withError(t *testing.T) {
	// Previously panicked with nil dereference.
	r := FromPtr[int](nil, errSentinel)
	if r.IsOk() {
		t.Fatal("FromPtr(nil, err) should not be ok")
	}
	if !errors.Is(r.Error(), errSentinel) {
		t.Fatalf("expected sentinel error, got %v", r.Error())
	}
}

func TestFromPtr_nilPtr_nilError(t *testing.T) {
	// Degenerate case: nil pointer, no error — should synthesise an error.
	r := FromPtr[int](nil, nil)
	if r.IsOk() {
		t.Fatal("FromPtr(nil, nil) should not be ok")
	}
	if r.Error() == nil {
		t.Fatal("FromPtr(nil, nil) should produce a non-nil error")
	}
}

// ── Get ──────────────────────────────────────────────────────────────────────

func TestGet_onErrorResult_doesNotPanic(t *testing.T) {
	// Previously panicked with a nil pointer dereference.
	r := Error[string](errSentinel)
	v, ok := r.Get()
	if ok {
		t.Fatal("Get() on error result should return false")
	}
	if v != "" {
		t.Fatalf("Get() on error result should return zero value, got %q", v)
	}
}

// ── OrElse ───────────────────────────────────────────────────────────────────

func TestOrElse(t *testing.T) {
	t.Run("ok result returns value", func(t *testing.T) {
		r := Ok("hello")
		got := r.OrElse("fallback")
		if got != "hello" {
			t.Fatalf("expected 'hello', got %q", got)
		}
	})
	t.Run("error result returns fallback", func(t *testing.T) {
		r := Error[string](errSentinel)
		got := r.OrElse("fallback")
		if got != "fallback" {
			t.Fatalf("expected 'fallback', got %q", got)
		}
	})
}

// ── OrElseGet ────────────────────────────────────────────────────────────────

func TestOrElseGet(t *testing.T) {
	t.Run("ok result does not call supplier", func(t *testing.T) {
		called := false
		r := Ok(1)
		got := r.OrElseGet(func() int { called = true; return 99 })
		if called {
			t.Fatal("supplier should not be called for an ok result")
		}
		if got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})
	t.Run("error result calls supplier", func(t *testing.T) {
		r := Error[int](errSentinel)
		got := r.OrElseGet(func() int { return 99 })
		if got != 99 {
			t.Fatalf("expected 99, got %d", got)
		}
	})
}

// ── IfOk / IfError ───────────────────────────────────────────────────────────

func TestIfOk(t *testing.T) {
	t.Run("called for ok result", func(t *testing.T) {
		called := false
		Ok("yes").IfOk(func(v string) { called = true })
		if !called {
			t.Fatal("IfOk consumer should be called on ok result")
		}
	})
	t.Run("not called for error result", func(t *testing.T) {
		called := false
		Error[string](errSentinel).IfOk(func(v string) { called = true })
		if called {
			t.Fatal("IfOk consumer should not be called on error result")
		}
	})
}

func TestIfError(t *testing.T) {
	t.Run("called for error result", func(t *testing.T) {
		called := false
		Error[int](errSentinel).IfError(func(e error) { called = true })
		if !called {
			t.Fatal("IfError consumer should be called on error result")
		}
	})
	t.Run("not called for ok result", func(t *testing.T) {
		called := false
		Ok(0).IfError(func(e error) { called = true })
		if called {
			t.Fatal("IfError consumer should not be called on ok result")
		}
	})
}

// ── String ───────────────────────────────────────────────────────────────────

func TestString(t *testing.T) {
	t.Run("ok result formats value", func(t *testing.T) {
		r := Ok(42)
		if r.String() != "42" {
			t.Fatalf("expected '42', got %q", r.String())
		}
	})
	t.Run("error result formats error message", func(t *testing.T) {
		r := Error[int](errSentinel)
		if r.String() != errSentinel.Error() {
			t.Fatalf("expected %q, got %q", errSentinel.Error(), r.String())
		}
	})
}

// ── ResultHolder interface satisfaction ──────────────────────────────────────

func TestResultHolder_interfaceSatisfaction(t *testing.T) {
	// Result[T] must implement ResultHolder[T] without any embedded field.
	var _ ResultHolder[int] = Ok(1)
	var _ ResultHolder[int] = Error[int](errSentinel)
}

// ── Try ──────────────────────────────────────────────────────────────────────

func TestTry_nilError_callsSuccess(t *testing.T) {
	called := false
	Try(nil, func() { called = true })
	if !called {
		t.Fatal("Try with nil error should call success")
	}
}

func TestTry_nonNilError_skipsSuccess(t *testing.T) {
	called := false
	Try(errSentinel, func() { called = true })
	if called {
		t.Fatal("Try with non-nil error should not call success")
	}
}

func TestTry_catchHandlerReceivesError(t *testing.T) {
	var caught error
	Try(errSentinel, func() {}, func(e error) bool {
		caught = e
		return true
	})
	if !errors.Is(caught, errSentinel) {
		t.Fatalf("expected sentinel error in catch, got %v", caught)
	}
}

func TestTry_catchHandlerReturnsFalse_continuesChain(t *testing.T) {
	calls := 0
	Try(errSentinel, func() {},
		func(e error) bool { calls++; return false }, // does not handle → next
		func(e error) bool { calls++; return true },  // handles
	)
	if calls != 2 {
		t.Fatalf("expected 2 catch handlers to be called, got %d", calls)
	}
}

func TestTry_catchHandlerReturnsTrue_stopsChain(t *testing.T) {
	calls := 0
	Try(errSentinel, func() {},
		func(e error) bool { calls++; return true }, // handles → stop
		func(e error) bool { calls++; return true }, // should not be reached
	)
	if calls != 1 {
		t.Fatalf("expected only 1 catch handler to be called, got %d", calls)
	}
}

// ── Zero-value guard ─────────────────────────────────────────────────────────

func TestZeroValue_isNotOk(t *testing.T) {
	// A zero-value Result{} has err==nil but value==nil.
	// IsOk() returns true for zero-value (err==nil), but Get() safely
	// returns (zero, false) because value is nil — this is documented behaviour.
	var r Result[int]
	// IsOk is true (err is nil) but Get signals no value.
	_, ok := r.Get()
	if ok {
		t.Fatal("Get() on zero-value Result should return false (value is nil)")
	}
}

// ── Table-driven: multiple types ─────────────────────────────────────────────

func TestOk_multipleTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		r := Ok("world")
		v, ok := r.Get()
		if !ok || v != "world" {
			t.Fatalf("expected ('world', true), got (%q, %v)", v, ok)
		}
	})
	t.Run("float64", func(t *testing.T) {
		r := Ok(3.14)
		v, ok := r.Get()
		if !ok || v != 3.14 {
			t.Fatalf("expected (3.14, true), got (%v, %v)", v, ok)
		}
	})
	t.Run("struct", func(t *testing.T) {
		type Point struct{ X, Y int }
		r := Ok(Point{1, 2})
		v, ok := r.Get()
		if !ok || v.X != 1 || v.Y != 2 {
			t.Fatalf("unexpected value %+v", v)
		}
	})
	t.Run("slice", func(t *testing.T) {
		r := Ok([]int{1, 2, 3})
		v, ok := r.Get()
		if !ok || len(v) != 3 {
			t.Fatalf("unexpected value %v", v)
		}
	})
}

// ── Stringer on wrapped error ────────────────────────────────────────────────

func TestString_wrappedError(t *testing.T) {
	wrapped := fmt.Errorf("outer: %w", errSentinel)
	r := Error[bool](wrapped)
	if r.String() != wrapped.Error() {
		t.Fatalf("expected %q, got %q", wrapped.Error(), r.String())
	}
}
