package cli

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/ak-repo/tcopy/internal/runner"
)

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name             string
		args             []string
		result           runner.Result
		copyErr          error
		code             int
		executed, copied bool
	}{
		{name: "help", args: []string{"--help"}},
		{name: "version", args: []string{"--version"}},
		{name: "empty", code: 2},
		{name: "separator only", args: []string{"--"}, code: 2},
		{name: "unknown flag", args: []string{"--bad"}, code: 2},
		{name: "success", args: []string{"echo", "hi"}, result: runner.Result{Started: true, Output: []byte("hi")}, executed: true, copied: true},
		{name: "separator", args: []string{"--", "-command"}, result: runner.Result{Started: true}, executed: true, copied: true},
		{name: "clipboard failure", args: []string{"true"}, result: runner.Result{Started: true}, copyErr: errors.New("clipboard failed"), code: 1, executed: true, copied: true},
		{name: "child precedence", args: []string{"false"}, result: runner.Result{Started: true, Code: 7, Err: errors.New("exit 7")}, copyErr: errors.New("clipboard failed"), code: 7, executed: true, copied: true},
		{name: "child 126", args: []string{"child"}, result: runner.Result{Started: true, Code: 126, Err: errors.New("exit 126")}, code: 126, executed: true, copied: true},
		{name: "child 127", args: []string{"child"}, result: runner.Result{Started: true, Code: 127, Err: errors.New("exit 127")}, code: 127, executed: true, copied: true},
		{name: "start failure", args: []string{"missing"}, result: runner.Result{Code: 127, Err: errors.New("missing")}, code: 127, executed: true},
		{name: "overflow", args: []string{"command"}, result: runner.Result{Started: true, Overflow: true}, code: 1, executed: true},
		{name: "interrupted", args: []string{"command"}, result: runner.Result{Started: true, Interrupted: true, Code: 130}, code: 130, executed: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out, stderr bytes.Buffer
			executed, copied := false, false
			args := tt.args
			if len(args) > 0 && args[0] == "--" {
				args = args[1:]
			}
			code := run(tt.args, strings.NewReader("input"), &out, &stderr,
				func(got []string, in io.Reader, o, e io.Writer) runner.Result {
					executed = true
					if !reflect.DeepEqual(got, args) {
						t.Fatalf("args: %v", got)
					}
					data, _ := io.ReadAll(in)
					if string(data) != "input" {
						t.Fatal("stdin lost")
					}
					return tt.result
				}, func(source func() io.Reader) error {
					copied = true
					data, _ := io.ReadAll(source())
					if !bytes.HasPrefix(data, []byte("$ ")) || !bytes.HasSuffix(data, tt.result.Output) {
						t.Fatalf("transcript: %q", data)
					}
					return tt.copyErr
				})
			if code != tt.code || executed != tt.executed || copied != tt.copied {
				t.Fatalf("code=%d execute=%v copy=%v", code, executed, copied)
			}
			if tt.name == "version" && out.String() != "tcopy "+Version+"\n" {
				t.Fatal(out.String())
			}
		})
	}
}
