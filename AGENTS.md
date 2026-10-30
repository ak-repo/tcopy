# Repository Guidelines

## Project Structure & Module Organization

`tcopy` is a Linux command wrapper built with Go 1.27 or newer. Its module is `github.com/ak-repo/tcopy`.

- `cmd/tcopy/main.go`: executable entrypoint.
- `internal/cli`: argument parsing, version reporting, and exit-status coordination.
- `internal/runner`: process execution, streaming, bounded capture, and signals.
- `internal/format`: command quoting and transcript construction.
- `internal/clipboard`: Wayland/X11 backend selection and copying.
- Go tests live beside implementation files as `*_test.go`; `tests/install_test.sh` tests installation with local fixtures.
- `README.md` documents usage; `docs/tc-project-document.md` describes architecture and acceptance status. Release automation lives in `.github/workflows/` and `.goreleaser.yaml`.

## Build, Test, and Development Commands

- `go build -o tcopy ./cmd/tcopy`: build the local executable.
- `./tcopy pwd`: run a command and copy its transcript; requires a graphical session and `wl-copy` or `xclip`.
- `gofmt -w cmd internal`: format Go source and tests.
- `go test ./...`: run package tests.
- `go test -race ./...`: check tests for data races.
- `go vet ./...`: run static checks.
- `bash tests/install_test.sh`: verify valid and corrupted installation archives without networking.
- `goreleaser check`: validate release configuration.

CI also cross-compiles Linux amd64 and arm64 binaries.

## Coding Style & Naming Conventions

Use standard Go conventions and `gofmt` indentation. Keep package names short and lowercase; use descriptive Go identifiers and `TestBehavior` test names. Keep executable setup in `cmd/` and implementation in focused `internal/` packages. Preserve direct argument execution; shell syntax requires an explicit shell.

## Testing Guidelines

Use Go's standard `testing` package and table-driven cases where useful. Add regression tests for changed behavior, especially exit-status precedence, quoting, capture limits, backend fallback, and signal cleanup. Use fixtures or narrow injected dependencies instead of requiring a live clipboard. No numeric coverage threshold is configured. Clipboard persistence requires separate manual Wayland/X11 verification.

## Commit & Pull Request Guidelines

History uses Conventional Commit-style subjects such as `feat: ...` and `build: ...`; follow that pattern with concise, imperative descriptions. PRs should explain the problem, resulting behavior, and validation performed, link relevant issues, and update documentation when behavior changes. Include terminal examples for CLI changes. Run the applicable CI checks before requesting review.

## Security & Configuration

Transcripts may contain secrets. Avoid sensitive test fixtures and diagnostics. Preserve bounded capture, clipboard deadlines, process-group cleanup, and checksum verification when changing execution or installation behavior.
