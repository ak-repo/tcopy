package format

import (
	"io"
	"testing"
)

func TestTranscript(t *testing.T) {
	for _, tt := range []struct {
		args         []string
		output, want string
	}{
		{[]string{"pwd"}, "", "$ pwd\n"},
		{[]string{"echo", "", "a b", "a'b", "$HOME", "a|b"}, "one\ntwo\n", "$ echo '' 'a b' 'a'\\''b' '$HOME' 'a|b'\none\ntwo\n"},
		{[]string{"printf", "hello"}, "hello", "$ printf hello\nhello"},
	} {
		got, err := io.ReadAll(Transcript(tt.args, []byte(tt.output)))
		if err != nil || string(got) != tt.want {
			t.Fatalf("got %q, %v; want %q", got, err, tt.want)
		}
	}
}
