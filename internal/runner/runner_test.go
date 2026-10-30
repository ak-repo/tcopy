package runner

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	var o, e bytes.Buffer
	r := Run([]string{"sh", "-c", "printf hello; printf error >&2; exit 7"}, nil, &o, &e)
	if r.Code != 7 || o.String() != "hello" || e.String() != "error" {
		t.Fatal(r)
	}
}
