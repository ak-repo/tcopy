# tc — Terminal Command & Output Copier

**Project status:** Proposed / MVP development  
**Version target:** `v0.1.0`  
**Primary language:** Go  
**Platform (v0.1):** Linux (Wayland and X11)  
**License (proposed):** MIT

## 1. Overview

`tc` is a lightweight command-line utility that runs a command, displays its output in the terminal, and copies **the command plus its output** to the system clipboard. It is designed for developers sharing build logs, debugging information, command results, and reproducible examples.

### Example

```bash
tc pwd
```

Terminal:

```text
/home/ak/downloads
```

Clipboard:

```text
$ pwd
/home/ak/downloads
```

For multi-line output, the clipboard retains the line breaks. Errors should be captured too.

## 2. Problem & Goals

Current terminal clipboard workflows often require piping output through tools such as `wl-copy` or `xclip`, which typically loses the original command and may hide output unless combined with `tee`.

**Goals**

- One-command execution and clipboard capture (`tc git status`).
- Show output in the terminal while capturing it.
- Include an accurately formatted representation of the invoked command.
- Capture both standard output and standard error.
- Return the executed command's exit status where practical.
- Support Linux Wayland and X11 without heavy runtime dependencies.
- Distribute a single compiled executable and publish open-source releases.

**Non-goals for v0.1**

- Replacing an interactive shell.
- Automatically capturing commands previously executed outside `tc`.
- Native support for interactive/full-screen applications (`vim`, `top`, etc.).
- Processing arbitrarily large or never-ending output without limits.
- Parsing shell metacharacters such as pipes or redirects without an explicit shell.

## 3. CLI Design

| Command | Expected behavior |
| --- | --- |
| `tc pwd` | Run `pwd`; show and copy command and path. |
| `tc git status` | Run `git status`; show and copy output. |
| `tc make run` | Show and copy build logs, including errors. |
| `tc --help` | Print supported usage and options. |
| `tc --version` | Print the installed version. |

**Proposed follow-on options (not v0.1 commitments)**

- `tc --only <command...>` — copy only output.
- `tc --md <command...>` — copy a Markdown fenced code block.
- `tc --save <file> -- <command...>` — save captured text to a file.
- `tc --last` — capture the previous shell command/output; **requires shell integration** and cannot be implemented reliably by a normal standalone executable alone.

### Format

Default clipboard format:

````text
$ <command and quoted arguments>
<combined output>
````

For empty output, copy the command header. Preserve final newlines and represent arguments safely, including spaces and quotes. A future formatting option may offer `command = output` for single-line results.

## 4. Command Name Conflict

**Important:** Linux commonly already includes an executable named `tc` from the `iproute2` suite, used for traffic control. Installing this utility as `tc` may shadow the existing system command depending on `PATH` order.

**Publishing recommendation:** Use a unique default executable name such as `tcopy`, and document an **opt-in** personal alias `alias tc='tcopy'` for users who prefer the short name. Users who need system traffic control must retain access to the original system `tc` command. If the repository is named `tc`, its binary can still be named `tcopy`.

## 5. Technology & Architecture

**Go** is recommended for v0.1: simple process management, fast compilation, cross-compilation, and a straightforward single-binary release. Rust is a viable alternative but not necessary for this scope.

### Components

1. **CLI parser** — validates flags and identifies the command plus arguments.
2. **Command runner** — invokes a process with `os/exec`, streams output to the terminal, captures it, and records exit status.
3. **Formatter** — builds clipboard text with command and output.
4. **Clipboard adapter** — invokes `wl-copy` on Wayland or `xclip -selection clipboard` on X11; reports missing backends clearly.
5. **Exit handling** — preserves child exit code when possible and handles interrupts and clipboard failures consistently.

### Suggested repository layout

```text
tc/
├── cmd/
│   └── tcopy/
│       └── main.go
├── internal/
│   ├── runner/
│   │   ├── runner.go
│   │   └── runner_test.go
│   ├── clipboard/
│   │   └── clipboard.go
│   └── format/
│       ├── format.go
│       └── format_test.go
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── go.mod
├── README.md
├── LICENSE
└── .gitignore
```

For an initial prototype, a single `main.go` is sufficient; modularize when tests and features justify it.

### Execution flow

```text
User: tcopy git status
        |
        v
Parse command and arguments
        |
        v
Spawn process and stream stdout/stderr
        |
        v
Capture bounded output and exit code
        |
        v
Format "$ git status\n<output>"
        |
        v
Copy via wl-copy / xclip
        |
        v
Exit with appropriate status
```

## 6. Implementation Notes

- **Do not** execute user arguments through `sh -c` by default. Use `exec.Command(program, args...)` to avoid accidental shell interpretation.
- A literal pipeline can be requested explicitly, e.g. `tcopy sh -c 'ls | grep go'`.
- Handle `--` to distinguish utility options from commands/options.
- Consider capturing stdout and stderr separately with clear semantics; perfect relative ordering across two OS streams is not guaranteed.
- When capturing both output streams, show output promptly rather than only after completion.
- For missing commands and other process-start errors, provide a useful error message and defined exit status.
- Handle `SIGINT`/`SIGTERM` and clean up child processes. A normal process wrapper does not automatically offer a full terminal/PTY experience.
- Set an output size limit or spool to a temporary file for long logs; avoid unbounded RAM growth.
- Treat clipboard failures separately from command failures and document exit-code precedence.
- Clipboard content may contain credentials, secrets, access tokens, or sensitive environment values. Do not upload content anywhere; make it clear when copying occurs.
- Linux clipboard commands require a usable graphical clipboard session; headless shells and remote SSH sessions may not have one.

## 7. Dependencies

**Build:** Go toolchain.  
**Runtime:** Linux plus one clipboard backend:

- Wayland: `wl-copy` from `wl-clipboard`
- X11: `xclip`

Ubuntu/Debian:

```bash
sudo apt update
sudo apt install wl-clipboard xclip
```

Clipboard selection should check both session variables and actual program availability, with a fallback where appropriate.

## 8. Development Setup

```bash
mkdir tc && cd tc
go mod init github.com/YOUR_USERNAME/tc
mkdir -p cmd/tcopy internal/{runner,clipboard,format}
```

After implementing `cmd/tcopy/main.go`:

```bash
go fmt ./...
go test ./...
go build -o tcopy ./cmd/tcopy
./tcopy pwd
./tcopy git status
```

Replace `YOUR_USERNAME` with the actual GitHub owner before publishing.

## 9. Testing & Acceptance Criteria

| Test case | Acceptance criterion |
| --- | --- |
| Simple command (`pwd`) | Terminal output remains visible; clipboard contains command and output. |
| Multi-word arguments (`echo hello world`) | Arguments are represented unambiguously. |
| Multi-line output | Line breaks are preserved. |
| Stdout + stderr | Both are captured and displayed. |
| Nonzero exit | Output is copied; command exit status is preserved or documented. |
| Missing executable | Clear error message, no silent success. |
| Empty output | Clipboard still contains command header. |
| No clipboard backend | Clear diagnostic; documented exit behavior. |
| Command with secrets | No networking or logs beyond explicitly requested terminal/clipboard behavior. |
| Large output | Bounded memory or documented limit. |
| Interrupt (`Ctrl+C`) | Child process does not remain running unintentionally. |

Use dependency injection around process and clipboard execution so automated tests do not require a live desktop clipboard.

## 10. Publishing Plan

### Stage A — Source repository

1. Add a clear `README.md` with examples, Linux dependencies, naming conflict note, limitations, and security caveats.
2. Add an MIT license (if chosen), `.gitignore`, tests, and contribution guidelines.
3. Enable CI for `go test ./...`, `go vet ./...`, and formatting checks.
4. Push the repository to GitHub.

Example with the GitHub CLI (after authenticating):

```bash
git init -b main
git add .
git commit -m "feat: initial terminal output copier"
gh repo create tc --public --source=. --remote=origin --push
```

### Stage B — Versioned release

```bash
git tag v0.1.0
git push origin v0.1.0
```

Publish release assets for at least `linux/amd64` and `linux/arm64`, including checksums. GitHub Actions or GoReleaser may automate compilation and release attachments.

Manual example:

```bash
mkdir -p dist
GOOS=linux GOARCH=amd64 go build -o dist/tcopy-linux-amd64 ./cmd/tcopy
GOOS=linux GOARCH=arm64 go build -o dist/tcopy-linux-arm64 ./cmd/tcopy
sha256sum dist/* > dist/checksums.txt
gh release create v0.1.0 dist/* --title "v0.1.0" --generate-notes
```

### Stage C — Installation instructions

**For Go users:**

```bash
go install github.com/YOUR_USERNAME/tc/cmd/tcopy@latest
```

The installed binary is named `tcopy`. Users must have `$(go env GOPATH)/bin` or `GOBIN` on their `PATH`.

**For non-Go users:** Download the matching GitHub Release binary, verify its checksum, rename it to `tcopy`, and place it in a directory on `PATH` (e.g., `~/.local/bin`). Avoid recommending an unaudited `curl | sh` installer as the default route.

**Optional short alias:**

```bash
alias tc='tcopy'
```

Document how to make the alias persistent in Bash, Zsh, or Fish, and how to invoke the original system `tc` when necessary.

## 11. Roadmap

### v0.1.0 — Linux MVP

- [ ] Execute basic non-interactive commands.
- [ ] Display and capture stdout/stderr.
- [ ] Copy command and output on Wayland/X11.
- [ ] Handle missing commands, nonzero exits, and clipboard failures.
- [ ] Add tests, help/version flags, and usage documentation.
- [ ] Publish Linux amd64/arm64 binaries.

### v0.2.0 — Quality of life

- [ ] `--only` and `--md` formatting modes.
- [ ] Output-size limits or disk-backed capture.
- [ ] Improved signals and cancellation behavior.
- [ ] Configurable clipboard backend and better headless diagnostics.

### v0.3.0+ — Wider adoption

- [ ] macOS clipboard backend (`pbcopy`).
- [ ] Windows clipboard support.
- [ ] Shell completion and packaging (Homebrew, distro packages, etc.).
- [ ] Experimental `--last` via explicit shell hooks, with privacy implications documented.

## 12. Definition of Done for v0.1.0

A release is ready when a Linux user can install `tcopy`, run `tcopy git status`, see the output live, paste a faithful command-and-output transcript, understand any clipboard failures, and download a tested versioned binary or install via the Go toolchain.

---

**Project decision:** Go implementation, Linux-first, public repository named `tc`, recommended executable name `tcopy` to avoid collision with the system `tc` networking utility.
