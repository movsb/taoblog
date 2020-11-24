package canonical

import "testing"

func TestIsValidPath(t *testing.T) {
	cases := []struct {
		Path  string
		Valid bool
	}{
		{``, false},
		{`file`, false},
		{`/`, true},
		{`/.`, false},
		{`/..`, false},
		{`/abc/.`, false},
		{`/abc/..`, false},
		{`/abc/./`, false},
		{`/abc/../`, false},
		{`/.gitignore`, true},
		{`/..xxx`, true},
		{`/abc/./def/`, false},
		{`/abc/../def/`, false},
	}
	for _, c := range cases {
		if isValidPath(c.Path) != c.Valid {
			t.Errorf(`fail: %s,%v`, c.Path, c.Valid)
		}
	}
}
