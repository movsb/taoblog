package main

import (
	_ "embed"
	"os"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
)

type Color struct {
	Name string `yaml:"name"`
	Hex  string `yaml:"hex"`
}

func main() {
	var (
		input  = os.Args[1]
		output = os.Args[2]
	)

	fpIn := utils.Must1(os.Open(input))
	defer fpIn.Close()

	var cs []Color
	utils.Must(yaml.NewDecoder(fpIn).Decode(&cs))
	fp := utils.Must1(os.Create(output))
	defer fp.Close()

	for i := range cs {
		cs[i].Name = strings.ToLower(cs[i].Name)
	}

	t := template.Must(template.New(`palette`).Parse(`
.color {
	{{- range . }}
	&.fg-{{.Name}} { color: {{ .Hex }}};
	&.bg-{{.Name}} { background-color: {{ .Hex }}};
	{{- end }}
}
`))

	utils.Must(t.Execute(fp, cs))
}
