# tcopy — Terminal Command & Output Copier

**Status:** Linux MVP implemented; release and desktop acceptance require verification.
**Release target:** v0.1.0
**Language:** Go (declared minimum: 1.27)
**Platforms:** Linux amd64 and arm64
**License:** MIT

## Purpose and current scope

`tcopy` runs a command, displays stdout/stderr live, and copies a command-and-output transcript for sharing logs and reproducible examples. It uses a single compiled executable plus an external desktop clipboard tool. The executable and repository are named `tcopy`; the Go module is `github.com/ak-repo/tcopy`.

The project document's historical filename is retained. The original `tc` naming proposal has been superseded to avoid shadowing Linux traffic control. No shell replacement, automatic capture of prior commands, PTY support, or cross-platform clipboard support is included.

Installation, usage, dependencies, limits, exit statuses, and troubleshooting are maintained in [README.md](../README.md).

## Architecture and interfaces

```text
cmd/tcopy → CLI → runner → formatter → clipboard adapter
                    │                     │
              terminal streams      wl-copy / xclip
```

- **Entrypoint:** passes process arguments and standard streams to `cli.Run`, then exits with its status.
- **CLI:** accepts `--help`, `--version`, and an optional `--` separator. The first command and all remaining arguments execute directly. Internal execution/copy functions can be substituted in tests; the public CLI syntax is unchanged.
- **Runner:** uses `os/exec`, inherited environment/directory/stdin, and a separate child process group. Stdout and stderr retain their destinations while a synchronized capture combines up to 10 MiB. Results distinguish successful process start, exit code, execution error, overflow, and interruption.
- **Formatter:** returns a reader containing `$ <quoted arguments>\n` followed by original captured bytes. Empty output preserves the header; empty or unsafe arguments are single-quoted, with embedded quotes escaped. Cross-stream ordering is approximate.
- **Clipboard:** accepts a fresh-reader factory so fallback receives the complete transcript. It tries `wl-copy` with `WAYLAND_DISPLAY`, then `xclip -selection clipboard` with `DISPLAY`, checking executable availability. Each attempt has a five-second context deadline and bounded pipe cleanup. Failed attempts are combined into a diagnostic.

Execution failures do not prevent copying a started child's transcript, including when it returns 126 or 127. Start failures skip copying. Overflow and received SIGINT/SIGTERM skip copying. Child failure status takes precedence; otherwise capture/clipboard failures return 1, usage errors return 2. Missing executables return 127; other start errors return 126.

SIGINT/SIGTERM forward to the child group, followed by forced cleanup after two seconds when necessary. A child terminated by a signal returns `128 + signal`. This is process wrapping without a PTY; interactive applications are unsupported.

## Development and distribution

Use the existing checkout rather than initializing a new module. Build with `go build -o tcopy ./cmd/tcopy`; development builds report `dev`. GoReleaser injects the release tag through the CLI's version variable.

Pull-request and main/master branch CI checks formatting, race tests, vet, both Linux builds, installer fixtures, and release configuration. The tag workflow repeats formatting, race tests, vet, installer fixtures, and configuration validation before publishing archives. No release availability is implied by these configuration files.

Release assets are `tcopy_<version>_linux_<arch>.tar.gz` and `checksums.txt`. Archives include README and LICENSE. The installer selects amd64/arm64, verifies the downloaded checksum, and installs into the configured directory. Native packages are deferred pending maintainer metadata and package testing.

Transcripts can contain credentials or other sensitive data. tcopy neither uploads nor persists them itself. Commands and clipboard managers can independently use networking or retain history.

## Acceptance status and verification

| Capability | Implementation / verification |
| --- | --- |
| Direct command execution, inherited stdin | Implemented; automated tests |
| Live stdout/stderr and bounded combined capture | Implemented; runner tests, including limit boundaries |
| Quoted arguments, empty/multiline output, final newlines | Implemented; formatter tests |
| Usage, version, separator, exit precedence | Implemented; CLI tests |
| Missing/unexecutable command and child 126/127 | Implemented; runner and CLI tests |
| Backend availability, preference, fallback, transcript replay | Implemented; local executable fixtures |
| Clipboard timeout | Implemented; short-deadline fixture test |
| SIGINT/SIGTERM and forced termination | Implemented; isolated subprocess tests |
| Installer valid/corrupt archives | Local release fixtures; no network required |
| Linux release builds and configuration | Automated CI checks; publishing requires a version tag |
| Real desktop clipboard persistence | Pending manual Wayland and X11 checks |
| Published versioned binaries / Go installation | Verify against an actual published release before marking complete |

Before release, run the README development checks and build both targets. On both Wayland and X11, run `tcopy pwd`, a command emitting stdout/stderr, an empty-output command, and a failing command; paste into another application after tcopy exits. Verify transcript content, continued clipboard ownership, terminal output, and exit status. Confirm headless diagnostics and interrupt cleanup. Inspect release assets and test installation from the published archives before declaring v0.1.0 complete.

## Roadmap

- **v0.1.0:** finish desktop acceptance, publish tested amd64/arm64 archives, and verify installation.
- **Quality of life:** consider `--only`, `--md`, saving output, configurable clipboard selection, and disk-backed capture. The current MVP already has output limits and signal handling.
- **Later:** macOS/Windows clipboard adapters, completion, native packages, and explicit shell integration for previous-command capture. `--last` requires shell hooks and a separate privacy design.

These are future proposals, not supported flags or release commitments.
