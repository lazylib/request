// Package request provides a small generic helper for sending HTTP requests
// and decoding JSON responses.
package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Options describes a single HTTP request.
//
// Body can be any value that json.Marshal can handle. If it is nil, an empty
// request body is sent. If it is already a []byte or io.Reader, it is used
// as-is (without re-encoding).
//
// Headers is merged on top of the Content-Type that the package sets for JSON
// bodies, so callers can override it.
type Options struct {
	Method  string
	Url     string
	Body    any
	Headers map[string]string
	Auth    Auth
}

// Auth is implemented by any value that can apply credentials to an
// outgoing http.Request (e.g. BasicAuth).
type Auth interface {
	apply(*http.Request)
}

// BasicAuth adds HTTP Basic credentials to the request.
type BasicAuth struct {
	Username string
	Password string
}

func (b BasicAuth) apply(r *http.Request) {
	r.SetBasicAuth(b.Username, b.Password)
}

// BearerAuth adds an Authorization: Bearer <token> header to the request.
type BearerAuth struct {
	Token string
}

func (b BearerAuth) apply(r *http.Request) {
	r.Header.Set("Authorization", "Bearer "+b.Token)
}

// Send performs opts and decodes the response body into *T.
//
// It returns an error if the request cannot be built or sent, if the server
// replies with a non-2xx status, or if the response body cannot be decoded
// as JSON.
func Send[T any](opts Options) (*T, error) {
	body, contentType, err := buildBody(opts.Body)
	if err != nil {
		return nil, fmt.Errorf("request: build body: %w", err)
	}

	req, err := http.NewRequest(opts.Method, opts.Url, body)
	if err != nil {
		return nil, fmt.Errorf("request: build request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}
	if opts.Auth != nil {
		opts.Auth.apply(req)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request: server responded with status %d", resp.StatusCode)
	}

	// Decode only if there is a body. json.Decode returns io.EOF on an
	// empty stream, which we treat as "nothing to decode" rather than an
	// error so callers don't have to special-case 204 / 304 responses.
	return decodeResponse[T](resp.Body)
}

func decodeResponse[T any](r io.Reader) (*T, error) {
	var data T
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("request: decode response: %w", err)
	}
	return &data, nil
}

// buildBody turns the user-supplied body into a reader + Content-Type.
//
//   - nil               -> nil reader, no Content-Type
//   - []byte / *bytes.Buffer / io.Reader -> used as raw bytes
//   - anything else     -> JSON-encoded with Content-Type: application/json
func buildBody(body any) (io.Reader, string, error) {
	switch v := body.(type) {
	case nil:
		return nil, "", nil
	case []byte:
		return bytes.NewReader(v), "", nil
	case *bytes.Buffer:
		if v == nil {
			return nil, "", nil
		}
		return v, "", nil
	case *bytes.Reader:
		if v == nil {
			return nil, "", nil
		}
		return v, "", nil
	case io.Reader:
		return v, "", nil
	case string:
		return bytes.NewReader([]byte(v)), "", nil
	default:
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(buf), "application/json", nil
	}
}
