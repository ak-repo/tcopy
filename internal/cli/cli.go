package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/ak-repo/tcopy/internal/clipboard"
	transcript "github.com/ak-repo/tcopy/internal/format"
	"github.com/ak-repo/tcopy/internal/runner"
)

const Usage = "Usage: tcopy [--] <command> [arguments...]\nRuns a non-interactive command and copies command plus output.\nOptions: --help, --version\nUse an explicit shell for pipes, redirects, aliases and built-ins.\n"

// Version is set by release builds using -ldflags.
var Version = "dev"

func Run(args []string, in io.Reader, out, errout io.Writer) int {
	return run(args, in, out, errout, runner.Run, clipboard.Copy)
}

func run(args []string, in io.Reader, out, errout io.Writer,
	execute func([]string, io.Reader, io.Writer, io.Writer) runner.Result,
	copyText func(func() io.Reader) error) int {
	if len(args) > 0 {
		switch args[0] {
		case "--help":
			fmt.Fprint(out, Usage)
			return 0
		case "--version":
			fmt.Fprintln(out, "tcopy "+Version)
			return 0
		case "--":
			args = args[1:]
		default:
			if strings.HasPrefix(args[0], "-") {
				fmt.Fprintln(errout, "tcopy: unknown option")
				return 2
			}
		}
	}
	if len(args) == 0 {
		fmt.Fprint(errout, Usage)
		return 2
	}
	r := execute(args, in, out, errout)
	failure := r.Err
	if r.Err != nil {
		fmt.Fprintln(errout, "tcopy:", r.Err)
	}
	if r.Overflow {
		failure = fmt.Errorf("output exceeds 10 MiB; clipboard copy skipped")
	}
	if r.Interrupted {
		failure = fmt.Errorf("interrupted; clipboard copy skipped")
	}
	if !r.Overflow && !r.Interrupted && r.Started {
		if e := copyText(func() io.Reader { return transcript.Transcript(args, r.Output) }); e != nil {
			failure = e
			fmt.Fprintln(errout, "tcopy:", e)
		}
	}
	if r.Overflow || r.Interrupted {
		fmt.Fprintln(errout, "tcopy:", failure)
	}
	if r.Code != 0 {
		return r.Code
	}
	if failure != nil {
		return 1
	}
	return 0
}
