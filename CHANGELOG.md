# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-07-27

### Added

- `Send[T any](Options) (*T, error)` — generic HTTP request helper with
  automatic JSON encoding/decoding.
- `Options` struct with `Method`, `Url`, `Body`, `Headers`, `Auth`.
- `BasicAuth` and `BearerAuth` built-in auth helpers.
- Pluggable `Auth` interface for custom auth schemes.
- Body handling for `nil`, `[]byte`, `*bytes.Buffer`, `*bytes.Reader`,
  `io.Reader`, `string`, and arbitrary JSON-marshalable values.
- Empty body / `204 No Content` returns `(nil, nil)` instead of error.
- Tests covering JSON POST, raw bytes, non-2xx errors, Basic auth, Bearer
  auth (YooKassa-style test case).
