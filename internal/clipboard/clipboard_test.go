package clipboard

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackends(t *testing.T) {
	for _, tt := range []struct{ name, wayland, x11, wl, xclip, want string }{
		{name: "headless", want: "no graphical clipboard session"},
		{name: "missing tool", wayland: "wayland-0", want: "wl-copy: not installed"},
		{name: "prefer wayland", wayland: "wayland-0", x11: ":0", wl: "/bin/cat > \"$CAPTURE\"", xclip: "exit 9"},
		{name: "fallback after failure", wayland: "wayland-0", x11: ":0", wl: "/bin/cat > /dev/null; exit 1", xclip: "/bin/cat > \"$CAPTURE\""},
		{name: "fallback missing tool", wayland: "wayland-0", x11: ":0", xclip: "/bin/cat > \"$CAPTURE\""},
		{name: "backend failure", x11: ":0", xclip: "exit 3", want: "exit status 3"},
		{name: "timeout", wayland: "wayland-0", wl: "exec /bin/sleep 10", want: "clipboard unavailable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			capture := filepath.Join(dir, "capture")
			t.Setenv("PATH", dir)
			t.Setenv("WAYLAND_DISPLAY", tt.wayland)
			t.Setenv("DISPLAY", tt.x11)
			t.Setenv("CAPTURE", capture)
			for name, script := range map[string]string{"wl-copy": tt.wl, "xclip": tt.xclip} {
				if script != "" {
					if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
						t.Fatal(err)
					}
				}
			}
			start := time.Now()
			err := copyWithTimeout(func() io.Reader { return strings.NewReader("complete transcript\n") }, 100*time.Millisecond)
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("error: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				got, e := os.ReadFile(capture)
				if e != nil || string(got) != "complete transcript\n" {
					t.Fatalf("capture %q: %v", got, e)
				}
			}
			if time.Since(start) > 2*time.Second {
				t.Fatal("timeout failed to bound execution")
			}
		})
	}
}
