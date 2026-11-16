# Contributing

Thanks for taking the time. Bug reports, ideas and pull requests are all
welcome.

## Getting started

```
git clone https://github.com/timzifer/fyne-breadcrumb
cd fyne-breadcrumb
go test ./...
go run ./cmd/demo
```

On Linux you need the Fyne build dependencies (OpenGL, X11 and Wayland
headers) for the demo; see the
[Fyne getting started guide](https://docs.fyne.io/started/).

## Before opening a pull request

- `gofmt -l .` prints nothing
- `go vet ./...` is clean
- `go test ./...` passes, and new behaviour comes with a test
- exported identifiers have doc comments that say *why*, not just *what*

CI runs the same checks on Linux, macOS and Windows with the race detector,
plus `golangci-lint`, and builds against both the minimum versions named in
`go.mod` (Go 1.22, Fyne v2.6) and the latest Fyne release. Please do not raise
either floor without a reason — say which API needs it in the pull request.

The library must keep at least 95 % statement coverage; CI fails below that
and shows the per function numbers in the job summary. The demo command is
not counted.

## Reporting a bug

Please include the Fyne version, the operating system and, if you can, a small
program that reproduces the problem.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
