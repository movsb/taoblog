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
table, tr td {
	min-width: 100px;
}
`,
			after: `article table,article tr td{min-width:100px;}`,
		},
	}
	for i, tc := range testCases {
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
