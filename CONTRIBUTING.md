# Contributing

Thanks for your interest in `lazylib/request`! This is a deliberately small
package, and the bar for new features is high. The goal is to keep the
public surface area as small as possible while staying useful.

## Ground rules

1. **No new dependencies.** The whole point of this package is "stdlib + generics".
2. **Small API surface.** Prefer extending `Options` or `Auth` over adding
   new functions.
3. **Backwards compatibility.** Anything that changes the signature of
   `Send` or `Options` is a breaking change and needs a major version bump.
4. **Tests for everything.** New behavior comes with a test in
   `request_test.go` (or an example in `./examples`).

## Local development

```bash
go test -race ./...
go vet ./...
gofmt -l .              # should print nothing
go test ./examples/...  # if you changed example code
```

## Submitting a PR

1. Fork the repo and create a topic branch.
2. Use the PR template.
3. Make sure CI is green.
4. Squash commits; write a message in the imperative mood
   (`add bearer auth timeout`, not `added …`).

## Reporting a bug

Use the **Bug report** issue template. Include a minimal reproduction
and your `go version`.

## Suggesting a feature

Open a **Feature request**. Be explicit about whether it is a breaking
change, and what you considered instead. Many features are a better fit
for a wrapper package on top of `request` than for `request` itself.
