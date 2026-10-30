package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var o, e bytes.Buffer
	r := Run([]string{"sh", "-c", "printf hello; printf error >&2; exit 7"}, nil, &o, &e)
	if r.Code != 7 || o.String() != "hello" || e.String() != "error" {
		t.Fatal(r)
	}
}

func TestCaptureAndStreaming(t *testing.T) {
	for _, n := range []int{Limit - 1, Limit, Limit + 1} {
		var dst bytes.Buffer
		cap := &capture{}
		s := stream{dst: &dst, cap: cap}
		data := bytes.Repeat([]byte("x"), n)
		if written, err := s.Write(data); err != nil || written != n {
			t.Fatalf("write: %d %v", written, err)
		}
		if _, err := s.Write([]byte("tail")); err != nil {
			t.Fatal(err)
		}
		if dst.Len() != n+4 || len(cap.data) != Limit || !cap.overflow {
			t.Fatalf("stream=%d capture=%d overflow=%v", dst.Len(), len(cap.data), cap.overflow)
		}
	}
	cap := &capture{}
	s := stream{dst: io.Discard, cap: cap}
	s.Write(bytes.Repeat([]byte("x"), Limit))
	if cap.overflow {
		t.Fatal("exact limit marked overflow")
	}
}

func TestRunInputAndStartFailures(t *testing.T) {
	var out bytes.Buffer
	r := Run([]string{"/bin/cat"}, strings.NewReader("input\n"), &out, io.Discard)
	if !r.Started || r.Code != 0 || out.String() != "input\n" || string(r.Output) != "input\n" {
		t.Fatalf("%+v", r)
	}
	t.Setenv("PATH", t.TempDir())
	r = Run([]string{"tcopy-missing-command"}, nil, io.Discard, io.Discard)
	if r.Started || r.Code != 127 || r.Err == nil {
		t.Fatalf("%+v", r)
	}
	r = Run([]string{filepath.Join(t.TempDir(), "missing")}, nil, io.Discard, io.Discard)
	if r.Started || r.Code != 127 || r.Err == nil {
		t.Fatalf("%+v", r)
	}
	file := filepath.Join(t.TempDir(), "unexecutable")
	if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}
	r = Run([]string{file}, nil, io.Discard, io.Discard)
	if r.Started || r.Code != 126 || r.Err == nil {
		t.Fatalf("%+v", r)
	}
}

// Isolate signal handling so a failure cannot terminate the test process.
func TestSignalHelper(t *testing.T) {
	if os.Getenv("TCOPY_SIGNAL_HELPER") != "1" {
		return
	}
	trap := "trap 'exit 42' INT TERM"
	if os.Getenv("TCOPY_IGNORE_SIGNAL") == "1" {
		trap = "trap '' INT TERM"
	}
	r := Run([]string{"/bin/sh", "-c", trap + `; printf ready > "$TCOPY_READY"; while :; do /bin/sleep 0.1; done`}, nil, io.Discard, io.Discard)
	if !r.Interrupted {
		os.Exit(99)
	}
	os.Exit(r.Code)
}

func TestSignalForwarding(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		for _, ignore := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/ignore=%v", sig, ignore), func(t *testing.T) {
				ready := filepath.Join(t.TempDir(), "ready")
				cmd := exec.Command(os.Args[0], "-test.run=^TestSignalHelper$")
				cmd.Env = append(os.Environ(), "TCOPY_SIGNAL_HELPER=1", "TCOPY_READY="+ready)
				if ignore {
					cmd.Env = append(cmd.Env, "TCOPY_IGNORE_SIGNAL=1")
				}
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = cmd.Process.Kill() })
				done := make(chan error, 1)
				go func() { done <- cmd.Wait() }()
				deadline := time.Now().Add(5 * time.Second)
				for {
					if _, err := os.Stat(ready); err == nil {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("child did not become ready")
					}
					time.Sleep(10 * time.Millisecond)
				}
				if err := cmd.Process.Signal(sig); err != nil {
					t.Fatal(err)
				}
				select {
				case err := <-done:
					want := 42
					if ignore {
						want = 137
					}
					var exit *exec.ExitError
					if !errors.As(err, &exit) || exit.ExitCode() != want {
						t.Fatalf("exit %v; want %d", err, want)
					}
				case <-time.After(7 * time.Second):
					t.Fatal("signal cleanup timed out")
				}
			})
		}
	}
}

// A terminal writer can fail even while the command exits successfully.
type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }

func TestShortWrite(t *testing.T) {
	s := stream{dst: shortWriter{}, cap: &capture{}}
	n, err := s.Write([]byte("output"))
	if n != 5 || !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("write: %d, %v", n, err)
	}
}
