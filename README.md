# tcopy
Linux command and output clipboard copier. Requires Go 1.27 and wl-copy (Wayland) or xclip (X11).

Build: `go build -o tcopy ./cmd/tcopy`

Usage: `tcopy git status`, `tcopy --help`, `tcopy --version`.
Use `tcopy sh -c 'ls | sort'` for shell syntax. Arguments otherwise execute directly. Environment, directory and stdin are inherited. Interactive applications are unsupported.

Output streams display live; relative stdout/stderr ordering is approximate. Capture is capped at 10 MiB; overflow skips copying while streaming continues. Clipboard attempts time out after five seconds. Wayland is preferred with X11 fallback when both session variables exist. Headless sessions report clipboard failure after executing.

Child failure statuses take precedence. Successful commands with clipboard/capture failures return 1; usage errors return 2. Signal exits use 128 + signal. SIGINT/SIGTERM forward to the child process group with forced cleanup after two seconds.

Clipboard content includes quoted command arguments and original output. Secrets may be copied automatically; nothing is uploaded. The tcopy name avoids collision with Linux traffic-control tc.

Run `go test -race ./...` and `go vet ./...` for validation. MIT licensed.
