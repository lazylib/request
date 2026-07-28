// Package request is a tiny, dependency-free Go helper for sending HTTP
// requests and decoding JSON responses into a typed value.
//
// It is designed for the common case: call an HTTP API, get JSON back, work
// with a struct. The whole surface area is a single generic function:
//
//	result, err := request.Send[MyResponse](request.Options{
//	    Method: http.MethodPost,
//	    Url:    "https://api.example.com/things",
//	    Body:   MyRequest{Name: "hello"},
//	    Headers: map[string]string{"X-Trace": "abc"},
//	    Auth:   request.BearerAuth{Token: token},
//	})
//
// The package handles JSON encoding of the body, JSON decoding of the
// response, non-2xx status codes, raw []byte / io.Reader payloads, and
// pluggable auth schemes. It has no third-party dependencies and works on
// Go 1.22+ (requires generics).
//
// # When to use this package
//
// Use it when stdlib net/http feels verbose and you don't need a full
// client (retries, connection pooling, middleware). If you only need
// "JSON in, JSON out", this is the entire API.
//
// # When NOT to use this package
//
// If you need retries, timeouts, circuit breakers, tracing, automatic
// rate-limit handling, or a pre-configured HTTP client, reach for
// net/http directly or a library such as hashicorp/go-retryablehttp.
package request
