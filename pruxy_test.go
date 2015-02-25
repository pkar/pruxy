package pruxy

import (
	"testing"
)

func TestRemoveTrailingSlash(t *testing.T) {
	var slashTests = []struct {
		in  string
		out string
	}{
		{"/ab/c/", "/ab/c"},
		{"/", "/"},
		{"/ab", "/ab"},
		{"", ""},
	}

	for _, tt := range slashTests {
		out := removeTrailingSlash(tt.in)
		if tt.out != out {
			t.Errorf("removeTrainlingSlash(%s) => %s, want %s", tt.in, out, tt.out)
		}
	}

}
