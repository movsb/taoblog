package scoped_css

import "testing"

func TestAddScope(t *testing.T) {
	testCases := []struct {
		scope  string
		before string
		after  string
	}{
		{
			scope: `article`,
			before: `
table, tr td, .a, #b, :c, ::d {
	min-width: 100px;
}
`,
			after: `article table,article tr td,article .a,article #b,article :c,article ::d{min-width:100px;}`,
		},
		{
			scope: `article`,
			before: `
@keyframes blinker {
    0% { opacity: 1; }
}`,
			after: `@keyframes blinker{0%{opacity:1;}}`,
		},
	}
	for i, tc := range testCases {
		if i == 0 {
			t.Log(`debug here`)
		}
		output, err := addScope(tc.before, tc.scope)
		if err != nil {
			t.Errorf(`Error %d: %v`, i, err)
			continue
		}
		if output != tc.after {
			t.Errorf("Not Equal: %d:\n%s\n%s", i, output, tc.after)
		}
	}
}
