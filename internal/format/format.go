package format

import (
	"io"
	"strings"
)

func Transcript(args []string, output []byte) io.Reader {
	q := make([]string, len(args))
	for i, s := range args {
		if s != "" && !strings.ContainsAny(s, " \t\n\r'\"\\$`;&|<>()*?[]{}!#~") {
			q[i] = s
		} else {
			q[i] = "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
		}
	}
	return io.MultiReader(strings.NewReader("$ "+strings.Join(q, " ")+"\n"), strings.NewReader(string(output)))
}
