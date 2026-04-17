# result

A generic `Result[T]` type for Go 1.18+ — represent success *or* failure
without the ambiguity of a bare `(value, error)` pair.

Inspired by Rust's `Result<T, E>` and Java's `Optional<T>`, this library
makes error paths explicit, chainable, and impossible to accidentally ignore.

---

## Installation

```bash
go get github.com/LCmaster/result
```

---

## Quick Start

```go
import "github.com/LCmaster/result"

// Wrap any standard (value, error) return pair
r := result.From(strconv.Atoi("42"))

// Check state
if r.IsOk() {
    v, _ := r.Get()
    fmt.Println("parsed:", v) // parsed: 42
}

// Provide a fallback
n := r.OrElse(0)

// React to each state
r.IfOk(func(v int)    { fmt.Println("ok:", v) })
r.IfError(func(e error) { fmt.Println("err:", e) })
```

---

## Constructors

| Function | When to use |
|---|---|
| `Ok(value)` | You have a definite success value |
| `Error[T](err)` | You have a definite error, no value |
| `From(value, err)` | Adapting a standard `(T, error)` return pair |
| `FromPtr(ptr, err)` | Adapting a `(*T, error)` return pair (e.g., DB queries) |

```go
// Explicit construction
ok  := result.Ok(42)
bad := result.Error[int](errors.New("not found"))

// Adapt existing Go functions directly — no boilerplate
r := result.From(os.Open("/etc/hosts"))
r := result.From(strconv.ParseFloat("3.14", 64))

// Adapt pointer-returning functions (repos, DB drivers, etc.)
row, err := db.FindUser(id)
r := result.FromPtr(row, err)
```

---

## Consuming a Result

### `Get() (T, bool)`

The fundamental accessor. Returns the value and `true` when ok, or the zero
value and `false` when an error. **Never panics.**

```go
if v, ok := r.Get(); ok {
    fmt.Println(v)
} else {
    fmt.Println("error:", r.Error())
}
```

### `OrElse(fallback T) T`

Returns the value when ok, otherwise returns `fallback`. Good for cheap
constant defaults.

```go
name := r.OrElse("anonymous")
```

### `OrElseGet(supplier func() T) T`

Like `OrElse`, but the fallback is computed lazily — the supplier is only
called when the result is an error. Prefer this when producing the fallback
is expensive.

```go
config := r.OrElseGet(func() Config { return loadDefaults() })
```

### `IfOk(consumer func(T))`

Runs a side-effectful callback only when the result is ok. A no-op on errors.

```go
r.IfOk(func(user User) {
    cache.Store(user.ID, user)
})
```

### `IfError(consumer func(error))`

Runs a side-effectful callback only when the result is an error. A no-op on success.

```go
r.IfError(func(e error) {
    log.Printf("failed to load user: %v", e)
})
```

---

## The `ResultHolder[T]` Interface

`Result[T]` implements `ResultHolder[T]`, an interface that exposes the full
read API. Use it in function signatures to decouple callers from the concrete
type — useful for testing and dependency injection.

```go
func renderUser(r result.ResultHolder[User]) string {
    if u, ok := r.Get(); ok {
        return u.Name
    }
    return "unknown"
}

// Works with both Ok and Error results
renderUser(result.Ok(user))
renderUser(result.Error[User](err))
```

---

## `Try` — Structured Error Handling

`Try` is a control-flow helper inspired by try/catch blocks. It runs the
success callback when there is no error, or walks through a variadic list of
catch handlers until one signals it handled the error.

```go
result.Try(err,
    func() {
        fmt.Println("operation succeeded")
    },
    func(e error) bool {
        if errors.Is(e, ErrNotFound) {
            fmt.Println("resource not found, using default")
            return true // stop — handled
        }
        return false // pass to the next handler
    },
    func(e error) bool {
        log.Printf("unhandled error: %v", e)
        return true
    },
)
```

> **Note:** `Try` does not provide a "finally" block. Unhandled errors are
> silently ignored if no catch handler returns `true`, so always add a
> catch-all handler for important operations.

---

## `String()` — Formatting

`Result[T]` implements `fmt.Stringer`. It renders the value when ok, or the
error message when an error, making it safe to pass directly to `fmt.Println`
and friends.

```go
fmt.Println(result.Ok(42))              // 42
fmt.Println(result.Error[int](io.EOF)) // EOF
```

---

## Design Notes

- **Invalid state is unrepresentable.** A `Result` holds a value *or* an error,
  never both. `From` discards the value when `err != nil`.
- **No panics in normal usage.** `Get()` returns `(zero, false)` instead of
  panicking when the result is an error.
- **`FromPtr` is nil-safe.** Passing a nil pointer always produces an error
  result, even if `err` is also nil.
- **Zero value is not valid.** Do not use `Result[T]{}` — always use a
  constructor.

---

## Requirements

- Go 1.18 or later (generics)
- No external dependencies

---

## License

[MIT](LICENSE)
