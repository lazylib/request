// Package request is a small, dependency-free Go toolkit for working with
// REST/JSON HTTP APIs. It grows by adding focused, composable helpers
// rather than shipping a single mega-client.
//
// The current public surface is intentionally small:
//
//   - [Send] — generic JSON in / typed Go out for any HTTP call.
//   - [SendX] — the panicking variant of [Send], for cases where a
//     non-2xx response is a programmer error or process-fatal condition.
//   - [Options], [Auth], [BasicAuth], [BearerAuth] — the configuration
//     types both helpers build on.
//
// A typical call looks like:
//
//	result, err := request.Send[MyResponse](request.Options{
//	    Method: http.MethodPost,
//	    Url:    "https://api.example.com/things",
//	    Body:   MyRequest{Name: "hello"},
//	    Headers: map[string]string{"X-Trace": "abc"},
//	    Auth:   request.BearerAuth{Token: token},
//	})
//
// JSON encoding of the body, JSON decoding of the response, non-2xx
// status codes, raw []byte / io.Reader payloads, and pluggable auth
// schemes are all handled by [Send]. The package has no third-party
// dependencies and works on Go 1.22+ (requires generics).
//
// # When to use this package
//
// Use it when stdlib net/http feels verbose and you want typed JSON
// responses without a full client (retries, connection pooling,
// middleware). [Send] covers "JSON in, typed Go out"; [SendX] covers
// the same case in a panic-on-failure form.
//
// # When NOT to use this package
//
// If you need retries, timeouts, circuit breakers, tracing, automatic
// rate-limit handling, or a pre-configured HTTP client, reach for
// net/http directly or a library such as hashicorp/go-retryablehttp.
package request
