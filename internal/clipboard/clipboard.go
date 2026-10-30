package clipboard

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Copy(source func() io.Reader) error {
	return copyWithTimeout(source, 5*time.Second)
}

func copyWithTimeout(source func() io.Reader, timeout time.Duration) error {
	var failures []string
	for _, b := range []struct {
		env, name string
		args      []string
	}{
		{"WAYLAND_DISPLAY", "wl-copy", nil}, {"DISPLAY", "xclip", []string{"-selection", "clipboard"}},
	} {
		if os.Getenv(b.env) == "" {
			continue
		}
		if _, err := exec.LookPath(b.name); err != nil {
			failures = append(failures, b.name+": not installed")
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		c := exec.CommandContext(ctx, b.name, b.args...)
		c.Stdin = source()
		c.WaitDelay = timeout
		err := c.Run()
		cancel()
		if err == nil {
			return nil
		}
		failures = append(failures, b.name+": "+err.Error())
	}
	if len(failures) == 0 {
		return fmt.Errorf("no graphical clipboard session: set WAYLAND_DISPLAY or DISPLAY in a desktop session")
	}
	return fmt.Errorf("clipboard unavailable (%s); install wl-clipboard or xclip", strings.Join(failures, "; "))
}
