package yaml_test

import (
	"fmt"
	"html"
	"os"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	test_utils "github.com/movsb/taoblog/modules/utils/test"
	"github.com/movsb/taoblog/service/modules/renderers/tables/yaml"
)

func TestAll(t *testing.T) {
	type TestCase struct {
		Description string     `yaml:"description"`
		Table       yaml.Table `yaml:"table"`
		HTML        string     `yaml:"html"`
	}
	testCases := test_utils.MustLoadCasesFromYaml[TestCase](`testdata/tests.yaml`)

	const name = `test.html`
	fp := utils.Must1(os.Create(name))
	defer fp.Close()
	fp.WriteString(`
	<!DOCTYPE html>
<html>
<head>
<style>
table, th, td {
	border: 1px solid gray;
	border-collapse: collapse;
}
th, td {
	padding: .5em 1em;
}

.bold       { font-weight: bold; }
.italic     { font-style: italic; }
.underline  { text-decoration-line: underline; }
.strike     { text-decoration-line: line-through; }
.code       { font-family: monospace; }
.ins        { color: green; }
.del        { color: red; }
.kbd        { font-family: monospace; }
.left       { text-align: left; }
.center     { text-align: center; }
.right      { text-align: right; }
.pre        { white-space: pre; }

</style>
</head>
<body>
`)
	defer fp.WriteString(`
</body>
</html>
`)
	for _, tc := range testCases {
		buf := strings.Builder{}
		utils.Must(tc.Table.Render(&buf))
		h := buf.String()
		if h != tc.HTML {
			fmt.Printf("%s:\nwant:\n%s\ngot:\n%s", tc.Description, tc.HTML, h)
			t.Error(`⬆️⬆️⬆️`)
			continue
		}
		fmt.Fprintf(fp, `<h2>%s</h2>`, html.EscapeString(tc.Description))
		fp.WriteString(buf.String())
	}
}
