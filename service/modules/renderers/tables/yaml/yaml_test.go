package yaml_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	goyaml "github.com/goccy/go-yaml"
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
	for _, tc := range testCases {
		buf := strings.Builder{}
		utils.Must(tc.Table.Render(&buf))
		h := buf.String()
		if h != tc.HTML {
			fmt.Printf("%s:\nwant:\n%s\ngot:\n%s", tc.Description, tc.HTML, h)
			t.Error(`⬆️⬆️⬆️`)
		}
	}
}

func TestRender(t *testing.T) {
	const table = `
rows:
  - cols:
      - 1
      - text: text content
        formats: [bold,italic,right]
      - 3
  - cols:
      - 1
      - text: text content / text content / 33234
        formats: [ins]
      - 3
  - cols:
      - 1
      - text: 2
        colspan: 2
  - - text: 4
      rowspan: 2
    - 5
    - 6
  - [8,9]
`

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
	padding: 1em;
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

</style>
</head>
<body>
`)

	tab := yaml.Table{}
	utils.Must(goyaml.Unmarshal([]byte(table), &tab))
	buf := strings.Builder{}
	utils.Must(tab.Render(&buf))
	fp.WriteString(buf.String())

	fp.WriteString(`
</body>
</html>
`)
}
