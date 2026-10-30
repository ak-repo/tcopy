package runner

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const Limit = 10 * 1024 * 1024

type Result struct {
	Output                []byte
	Code                  int
	Err                   error
	Overflow, Interrupted bool
}
type capture struct {
	sync.Mutex
	data     []byte
	overflow bool
}
type stream struct {
	dst io.Writer
	cap *capture
}

func (s stream) Write(p []byte) (int, error) {
	s.cap.Lock()
	defer s.cap.Unlock()
	n := len(p)
	remaining := Limit - len(s.cap.data)
	if n > remaining {
		s.cap.overflow = true
		s.cap.data = append(s.cap.data, p[:remaining]...)
	} else {
		s.cap.data = append(s.cap.data, p...)
	}
	_, err := s.dst.Write(p)
	return n, err
}
func Run(args []string, in io.Reader, out, stderr io.Writer) Result {
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = in
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.WaitDelay = 2 * time.Second
	b := &capture{}
	c.Stdout = stream{out, b}
	c.Stderr = stream{stderr, b}
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	if err := c.Start(); err != nil {
		code := 126
		if _, ok := err.(*exec.Error); ok {
			code = 127
		}
		return Result{Code: code, Err: err}
	}
	done := make(chan struct{})
	stopped := make(chan bool, 1)
	go func() {
		interrupted := false
		var timer <-chan time.Time
		for {
			select {
			case sig := <-signals:
				interrupted = true
				_ = syscall.Kill(-c.Process.Pid, sig.(syscall.Signal))
				timer = time.After(2 * time.Second)
			case <-timer:
				_ = syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
				timer = nil
			case <-done:
				stopped <- interrupted
				return
			}
		}
	}()
	err := c.Wait()
	close(done)
	interrupted := <-stopped
	code := 0
	if err != nil {
		code = 1
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
			if st, ok := e.Sys().(syscall.WaitStatus); ok && st.Signaled() {
				code = 128 + int(st.Signal())
			}
		}
	}
	if interrupted {
		_ = syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
	}
	return Result{b.data, code, err, b.overflow, interrupted}
}
