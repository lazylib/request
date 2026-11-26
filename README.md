# lazylib/request

[![CI](https://github.com/lazylib/request/actions/workflows/ci.yml/badge.svg)](https://github.com/lazylib/request/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lazylib/request.svg)](https://pkg.go.dev/github.com/lazylib/request)
[![Go Report Card](https://goreportcard.com/badge/github.com/lazylib/request)](https://goreportcard.com/report/github.com/lazylib/request)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/lazylib/request)](go.mod)

> A small, dependency-free Go toolkit for REST/JSON HTTP APIs. Typed `Send` and `SendX` helpers, pluggable auth, raw-body support — composed from focused building blocks rather than a mega-client.

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

`Send` for the err-returning case, `SendX` for the panic-on-failure case. Same `Options`, same JSON handling, same auth — pick the one that fits the call site.

---

## Why?

Every Go project that calls a JSON HTTP API eventually rewrites the same
helper: build a request, set headers, marshal the body, send it, check the
status, decode the response, return the error. `lazylib/request` ships
that helper — and a few siblings — battle-tested and zero-dependency, so
you can stop writing it again.

| What you want                       | What you write                                  |
| ----------------------------------- | ----------------------------------------------- |
| GET JSON, decode into struct        | `request.Send[T](Options{Method: GET, Url: …})` |
| Same as above, but panic on failure | `request.SendX[T](Options{…})`                  |
| Download a file                     | `request.Download(Options{…}, "file.bin")`      |
| One config for the whole app        | `api := request.New(Settings{…})`, then `api.Send[T]` |
| POST struct as JSON                 | pass `Body: myStruct`                           |
| POST raw bytes / stream             | pass `Body: []byte` or `io.Reader`              |
| HTTP Basic                          | `Auth: &BasicAuth{User, Pass}`                  |
| Bearer token                        | `Auth: BearerAuth{Token: …}`                    |
| Custom auth scheme                  | implement the `Auth` interface (2 lines)        |
| Custom headers                      | `Headers: map[string]string{…}`                 |

## Config — one client for the whole app

Applications that talk to one API usually need the same base URL, headers,
auth and TLS setup on every call. Build a `Config` once, keep it in your
own package-level variable (one per API, per environment, per region), and
use it everywhere:

```go
// Custom TLS certificate + default auth — configure once…
raw, err := os.ReadFile("certs/ca.pem")
if err != nil {
    log.Fatal(err)
}
pool := x509.NewCertPool()
if !pool.AppendCertsFromPEM(raw) {
    log.Fatal("bad CA certificate")
}

var api = request.New(request.Settings{
    BaseURL: "https://api.internal.example.com/v1",
    Auth:    request.BearerAuth{Token: os.Getenv("API_TOKEN")},
    Headers: map[string]string{"X-Api-Key": os.Getenv("API_KEY")},
    Client: &http.Client{
        Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}},
    },
})

// …then every call site shrinks to Method/Url/Body.
user, err := api.Send[User](request.Options{Method: http.MethodGet, Url: "/users/42"})
```

Need several environments? Just declare several configs — they are fully
independent:

```go
var (
    ru  = request.New(request.Settings{BaseURL: "https://ru.example.com", Auth: …})
    www = request.New(request.Settings{BaseURL: "https://www.example.com", Auth: …})
)

ru.Send[User](request.Options{Method: http.MethodGet, Url: "/users/1"})
www.Send[User](request.Options{Method: http.MethodGet, Url: "/users/1"})
```

**Methods:** `Config` exposes `Send[T]`, `SendX[T]`, `Download`,
`DownloadX` — the same semantics as the package-level helpers, with the
config's settings applied first.

**Resolution rules** (values set directly in `Options` always win):

- `Auth`, `Client` — used from `Settings` when the `Options` field is zero.
- `Url` — relative paths are joined onto `Settings.BaseURL` (extra slashes
  are collapsed); absolute URLs (containing `://`) bypass the base URL.
- `Headers` — merged: `Settings.Headers` first, then `Options.Headers`,
  which also overrides the automatic `Content-Type` for JSON bodies.

## Install

```bash
go get github.com/lazylib/request
```

Requires **Go 1.27+** (generic methods on Config).

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

### `SendX[T any](opts Options) *T`

The panicking variant of [`Send`](#sendt-anyopts-options-t-error). Calls
`Send[T]` and panics with the returned error if it is non-nil. Useful
when an HTTP failure is a programmer error or process-fatal condition
(similar to `regexp.MustCompile` or `template.Must`).

```go
config := request.SendX[Config](request.Options{
    Method: http.MethodGet,
    Url:    "https://api.example.com/config",
    Auth:   request.BearerAuth{Token: token},
})
// config is *Config, never nil, no error to check.
```

For most code paths, prefer [`Send`](#sendt-anyopts-options-t-error) and
return the error.

### `Options`

```go
type Options struct {
    Method  string             // "GET", "POST", ...
    Url     string             // absolute URL
    Body    any                // nil | struct/map | []byte | *bytes.Buffer | *bytes.Reader | io.Reader | string
    Headers map[string]string  // optional, merged on top of Content-Type for JSON bodies
    Auth    Auth               // optional, BasicAuth / BearerAuth / your own
    Client  *http.Client       // custom client
}
```

**Body handling:**

| Body type                       | Sent as              | Content-Type set   |
| ------------------------------- | -------------------- | ------------------ |
| `nil`                           | (empty)              | (none)             |
| `[]byte`                        | raw                  | (none)             |
| `*bytes.Buffer`/`*bytes.Reader` | raw                  | (none)             |
| `io.Reader`                     | raw                  | (none)             |
| `string`                        | raw                  | (none)             |
| anything else                   | `json.Marshal(body)` | `application/json` |

If you set `Headers["Content-Type"]`, it overrides the automatic value.

### `New` / `Config` / `Settings`

```go
type Settings struct {
    BaseURL string            // prepended to relative Url values
    Headers map[string]string // added to every request, overridden by Options.Headers
    Auth    Auth              // default auth, overridden by Options.Auth
    Client  *http.Client      // custom client (TLS, timeouts), overridden by Options.Client
}

func New(settings Settings) Config // build a reusable client

// Config methods — same semantics as the package-level helpers:
//   Send[T any](opts Options) (*T, error)
//   SendX[T any](opts Options) *T
//   Download(opts Options, to string) error
//   DownloadX(opts Options, to string)
```

See [Config](#config--one-client-for-the-whole-app) for a full example
and resolution rules.

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

Active. The toolkit grows by adding small, focused helpers on top of
the same `Options` and `Auth` primitives — `Send` and `SendX` today,
more building blocks over time. The public surface is designed to be
backwards-compatible across minor versions: new helpers are additive,
and existing call sites are not expected to change.

## Contributing

Bug reports and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) and
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Run `go test ./...` and
`go vet ./...` before submitting.

## License

[MIT](LICENSE).
