# Security Policy

## Supported versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

This package has a tiny API and is maintained on a best-effort basis.
Bug fixes and security fixes are released as new minor versions.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security
vulnerabilities. Instead, email the maintainers (see the commit history
for contact info) with:

- a description of the issue
- a minimal reproducer
- the impact you believe it has

We will acknowledge within 72 hours and aim to ship a fix within 7 days
for confirmed issues.

## What this package does NOT do

Out of scope for security reports (please don't file these):

- TLS / certificate validation — delegated to `crypto/tls` in the standard
  library.
- Retry storms / connection exhaustion — there is no built-in client; you
  can pass your own `*http.Client` via a future enhancement.
- Auth-scheme-specific bugs (e.g. "OAuth 1.0 signature is wrong") — this
  package ships only `BasicAuth` and `BearerAuth`. Custom auth is the
  caller's responsibility.
