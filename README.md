# tcopy

Run a Linux command, display its output live, and copy the command plus its stdout and stderr to your desktop clipboard. MIT licensed.

```sh
tcopy git status
tcopy sh -c 'ls | sort'
tcopy --help
tcopy --version
```

For `tcopy printf 'hello\n'`, the clipboard contains:

```text
$ printf 'hello\n'
hello
```

## Installation

Runtime requirements: Linux and a graphical clipboard session with `wl-copy` (Wayland) or `xclip` (X11). On Debian/Ubuntu:

```sh
sudo apt install wl-clipboard xclip
```

Build from this checkout with Go 1.27 or newer:

```sh
go build -o tcopy ./cmd/tcopy
./tcopy pwd
```

Once a versioned release is published, Go users can install with:

```sh
go install github.com/ak-repo/tcopy/cmd/tcopy@latest
```

Add `$(go env GOPATH)/bin` (or your configured `GOBIN`) to `PATH`.

Release configuration produces Linux amd64 and arm64 `.tar.gz` archives and `checksums.txt`. When a release is available, download both from [GitHub Releases](https://github.com/ak-repo/tcopy/releases), verify the selected archive with `sha256sum`, extract it, and install `tcopy` into `~/.local/bin` or another directory on `PATH`. For example, after downloading the amd64 v0.1.0 assets:

```sh
grep -F '  tcopy_0.1.0_linux_amd64.tar.gz' checksums.txt | sha256sum -c -
tar -xzf tcopy_0.1.0_linux_amd64.tar.gz
install -Dm755 tcopy "$HOME/.local/bin/tcopy"
```

Alternatively, inspect and run the repository's `install.sh`. It downloads and verifies the matching archive before installation. It requires Bash, `curl`, `uname`, `mktemp`, `grep`, `sha256sum`, `tar`, and `install`:

```sh
VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" bash install.sh
```

`VERSION` defaults to `v0.1.0`; `INSTALL_DIR` defaults to `~/.local/bin`. The requested release must already exist. Native DEB/RPM/Arch packages are deferred.

## Command behavior

- Arguments execute directly, without shell interpretation. Request pipes, redirects, shell built-ins, or aliases through an explicit shell; aliases require that shell's configuration.
- `tcopy [--] <command> [arguments...]` accepts `--help` and `--version` before the command. `--` ends tcopy option parsing; arguments after the command belong to that command.
- The child inherits the environment, working directory, and stdin. Full-screen and interactive applications are unsupported; tcopy does not allocate a PTY.
- Stdout and stderr display on their respective terminal streams and are combined in the clipboard transcript. Relative ordering across the two streams is approximate.
- Transcript arguments use shell-style quoting. Output bytes and final newlines are preserved; commands with no output still produce a command header.
- Capture is limited to 10 MiB. On overflow, terminal streaming continues and copying is skipped.
- SIGINT/SIGTERM are forwarded to the child process group. Forced cleanup follows after two seconds if needed; interrupted runs skip copying. Signal exits use `128 + signal`.

## Exit status and troubleshooting

| Situation | Status |
| --- | --- |
| Command and clipboard succeed | 0 |
| Child fails | Child status, including 126/127 if returned by a started command |
| Executable not found | 127 |
| Other command-start failure | 126 |
| Successful child, capture or clipboard failure | 1 |
| Missing command or unknown tcopy option | 2 |

Child failure takes precedence over clipboard or capture failure. A started command's output is copied even when it exits unsuccessfully, unless capture overflowed or tcopy received an interrupt.

Wayland is attempted first when `WAYLAND_DISPLAY` is set and `wl-copy` is installed. X11 is tried next when `DISPLAY` is set and `xclip` is installed, including after a Wayland failure. Each clipboard attempt has a five-second timeout, with bounded pipe cleanup. With both backends, total clipboard handling may take longer than five seconds.

A missing backend or unusable desktop session produces a diagnostic after the command executes. In headless/SSH sessions, use an actual accessible graphical session; setting a display variable alone does not establish one. Install the backend appropriate to that session.

Command arguments and output may contain secrets and are copied automatically. tcopy does not upload transcripts or keep a transcript log; the invoked command and desktop clipboard manager may have their own behavior.

The executable is named `tcopy` to avoid Linux's traffic-control `tc`. An optional personal alias is `alias tc='tcopy'`; this shadows the networking command in that shell.

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
bash tests/install_test.sh
goreleaser check
```

CI also checks Go formatting and builds both Linux architectures. Tests use local clipboard/installer fixtures; real clipboard persistence still requires Wayland and X11 desktop checks. Local builds report `tcopy dev`; release builds inject their tag into `--version`.

See [the project document](docs/tc-project-document.md) for architecture, acceptance status, and the roadmap.
