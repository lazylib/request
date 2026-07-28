# lazylib/request

[![CI](https://github.com/lazylib/request/actions/workflows/ci.yml/badge.svg)](https://github.com/lazylib/request/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lazylib/request.svg)](https://pkg.go.dev/github.com/lazylib/request)
[![Go Report Card](https://goreportcard.com/badge/github.com/lazylib/request)](https://goreportcard.com/report/github.com/lazylib/request)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/lazylib/request)](go.mod)

> One generic function for "JSON in, typed Go out" over HTTP. No dependencies. No boilerplate.

```go
result, err := request.Send[User](request.Options{
    Method:  http.MethodGet,
    Url:     "https://api.example.com/users/42",
    Headers: map[string]string{"Accept": "application/json"},
    Auth:    request.BearerAuth{Token: token},
})
if err != nil {
    return err
}
fmt.Println(result.Name)
```

That is the whole API. Drop it into any Go project that talks to JSON HTTP APIs.

---

## Why?

Every Go project that calls a JSON HTTP API eventually rewrites the same
helper: build a request, set headers, marshal the body, send it, check the
status, decode the response, return the error. `lazylib/request` is that
helper, battle-tested and zero-dependency.

| What you want                                  | What you write                                |
| ---------------------------------------------- | --------------------------------------------- |
| GET JSON, decode into struct                   | `request.Send[T](Options{Method: GET, Url: …})` |
| POST struct as JSON                            | pass `Body: myStruct`                         |
| POST raw bytes / stream                        | pass `Body: []byte` or `io.Reader`            |
| HTTP Basic                                     | `Auth: &BasicAuth{User, Pass}`                |
| Bearer token                                   | `Auth: BearerAuth{Token: …}`                  |
| Custom auth scheme                             | implement the `Auth` interface (2 lines)      |
| Custom headers                                 | `Headers: map[string]string{…}`               |
| Custom client / timeout / retries              | see [When NOT to use](#when-not-to-use)       |

## Install

```bash
go get github.com/lazylib/request
```

Requires **Go 1.22+** (uses generics).

## Quick start

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/lazylib/request"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    u, err := request.Send[User](request.Options{
        Method: http.MethodGet,
        Url:    "https://api.example.com/users/1",
        Auth:   request.BearerAuth{Token: "secret"},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(u.Name)
}
```

POST a JSON body, decode the JSON response:

```go
type CreateUserReq struct {
    Name string `json:"name"`
}
type CreateUserResp struct {
    ID int `json:"id"`
}

resp, err := request.Send[CreateUserResp](request.Options{
    Method: http.MethodPost,
    Url:    "https://api.example.com/users",
    Body:   CreateUserReq{Name: "Alice"},
    Auth:   request.BearerAuth{Token: "secret"},
})
```

POST raw bytes (e.g. pre-signed payload, file upload, webhook body):

```go
_, err := request.Send[struct{}](request.Options{
    Method: http.MethodPost,
    Url:    "https://api.example.com/hooks/42",
    Body:   []byte(`{"event":"ping"}`),
})
```

Send a request with no body and no response body (e.g. `204 No Content`):

```go
_, err := request.Send[struct{}](request.Options{
    Method: http.MethodDelete,
    Url:    "https://api.example.com/users/1",
    Auth:   request.BearerAuth{Token: "secret"},
})
```

HTTP Basic auth (e.g. payment gateways like YooKassa, Cloudflare R2, internal APIs):

```go
resp, err := request.Send[Payment](request.Options{
    Method: http.MethodPost,
    Url:    "https://api.yookassa.ru/v3/payments",
    Body:   payload,
    Auth:   &request.BasicAuth{Username: "shop-id", Password: "secret"},
})
```

## API

### `Send[T any](opts Options) (*T, error)`

Performs the request and decodes the response body into `*T`.

Returns an error if:

- the request cannot be built or sent (network error, invalid URL)
- the server replies with a non-2xx status — error message includes the status code
- the response body cannot be decoded as JSON
- for `204 No Content` / empty bodies, returns `(nil, nil)` — no need to special-case

### `Options`

```go
type Options struct {
    Method  string             // "GET", "POST", ...
    Url     string             // absolute URL
    Body    any                // nil | struct/map | []byte | *bytes.Buffer | *bytes.Reader | io.Reader | string
    Headers map[string]string  // optional, merged on top of Content-Type for JSON bodies
    Auth    Auth               // optional, BasicAuth / BearerAuth / your own
}
```

**Body handling:**

| Body type             | Sent as                          | Content-Type set     |
| --------------------- | -------------------------------- | -------------------- |
| `nil`                 | (empty)                          | (none)               |
| `[]byte`              | raw                              | (none)               |
| `*bytes.Buffer`/`*bytes.Reader` | raw                       | (none)               |
| `io.Reader`           | raw                              | (none)               |
| `string`              | raw                              | (none)               |
| anything else         | `json.Marshal(body)`             | `application/json`   |

If you set `Headers["Content-Type"]`, it overrides the automatic value.

### `Auth` interface

```go
type Auth interface {
    apply(*http.Request) // unexported
}
```

Implement this interface to plug in any auth scheme (HMAC, OAuth1, mTLS, signature v4, …):

```go
type MyAuth struct{ Key, Secret string }

func (a MyAuth) apply(r *http.Request) {
    r.Header.Set("X-Api-Key", a.Key)
    // …sign the request however you like
}

// Auth: MyAuth{Key: "k", Secret: "s"}
```

Built-in helpers:

- `BasicAuth{Username, Password string}` — `Authorization: Basic …`
- `BearerAuth{Token string}` — `Authorization: Bearer …`

## Comparison

### vs. stdlib `net/http`

```go
// stdlib — 10+ lines, easy to get wrong
req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)
resp, err := http.DefaultClient.Do(req)
if err != nil { return err }
defer resp.Body.Close()
if resp.StatusCode < 200 || resp.StatusCode >= 300 { return fmt.Errorf("status %d", resp.StatusCode) }
var out MyResponse
return json.NewDecoder(resp.Body).Decode(&out)
```

```go
// this package — 1 call
return request.Send[MyResponse](request.Options{
    Method: http.MethodPost, Url: url, Body: body,
    Auth: request.BearerAuth{Token: token},
})
```

### vs. `net/http` (Go 1.22+ has `http.NewRequestWithContext` — still verbose)

This package sits **on top of** `net/http`. It does not replace it. For
complex needs (timeouts, retries, middleware, connection pooling), use
`net/http` directly or a heavier client like
[`hashicorp/go-retryablehttp`](https://github.com/hashicorp/go-retryablehttp).

### vs. `resty`, `gentleman`, `req`, `gorequest`

Those are full HTTP clients with retries, middleware, and chainable
builders. This package is **deliberately tiny** — one function, one
struct, two auth types, zero dependencies. Use it when you want a
stdlib-feeling helper, not a framework.

## When NOT to use

You probably want `net/http` or `hashicorp/go-retryablehttp` instead if you need:

- retries with backoff
- per-request timeouts and a shared `*http.Client`
- request/response middleware, tracing, metrics
- OAuth2 token refresh flows
- multipart form uploads, streaming downloads
- WebSockets / SSE

## Examples

Runnable examples live in [`./examples`](./examples). They cover:

- [`examples/getjson`](./examples/getjson) — GET + decode JSON
- [`examples/postjson`](./examples/postjson) — POST struct as JSON
- [`examples/rawpayload`](./examples/rawpayload) — POST raw bytes
- [`examples/bearerauth`](./examples/bearerauth) — Bearer token
- [`examples/basicauth`](./examples/basicauth) — HTTP Basic

## Project status

Stable. Used in production. The API is small enough that breaking changes
are unlikely — anything new is added via the `Auth` interface and the
`Options` struct without touching existing call sites.

## Contributing

Bug reports and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) and
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Run `go test ./...` and
`go vet ./...` before submitting.

## License

[MIT](LICENSE).
