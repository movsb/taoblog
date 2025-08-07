package colors

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	"golang.org/x/net/html"
)

func TestShouldBeBlockElement(t *testing.T) {
	testCases := []struct {
		HTML    string
		IsBlock bool
	}{
		{
			HTML:    `<color/>`,
			IsBlock: true,
		},
		{
			HTML:    `<p><color/></p>`,
			IsBlock: false,
		},
		{
			HTML:    `<body><color></color></body>`,
			IsBlock: true,
		},
		{
			HTML:    `<li><color></color></li>`,
			IsBlock: false,
		},
		{
			HTML:    `<li>    <color></color>   </li>`,
			IsBlock: false,
		},
		{
			HTML:    `<li>a<color></color></li>`,
			IsBlock: false,
		},
		{
			HTML:    `<table><tr><td><color></color></td></tr></table>`,
			IsBlock: false,
		},
		{
			HTML:    `<td>a<color>b</color>c</td>`,
			IsBlock: false,
		},
		{
			HTML:    `<color> <p></p> </color>`,
			IsBlock: true,
		},
	}
	for _, tc := range testCases {
		gq, err := goquery.NewDocumentFromReader(strings.NewReader(tc.HTML))
		if err != nil {
			t.Error(err)
			continue
		}
		test := func() bool {
			return shouldBeBlockElement(gq.Find(`color`))
		}
		if test() != tc.IsBlock {
			t.Errorf(`不匹配：%s, %v`, tc.HTML, tc.IsBlock)
			test() // for debug
			html.Render(os.Stdout, gq.Nodes[0])
			fmt.Println()
		}
	}
}

func TestTransform(t *testing.T) {
	testCases := []struct {
		Before string
		After  string
	}{
		{
			Before: `<color></color>`,
			After:  `<div></div>`,
		},
		{
			Before: `<color red></color>`,
			After:  `<div class="color fg-red"></div>`,
		},
		{
			Before: `<p><color :blue></color></p>`,
			After:  `<p><span class="color bg-blue"></span></p>`,
		},
		{
			Before: `<p><color red:blue></color></p>`,
			After:  `<p><span class="color fg-red bg-blue"></span></p>`,
		},
		{
			Before: `<color fg="red" bg="blue"></color>`,
			After:  `<div style="color:red;background-color:blue;"></div>`,
		},
		{
			Before: `<color fg="rgb(1,1,1)" bg="hsl(330,100%,50%)"></color>`,
			After:  `<div style="color:rgb(1,1,1);background-color:hsl(330,100%,50%);"></div>`,
		},
	}
	for _, tc := range testCases {
		gq, err := goquery.NewDocumentFromReader(strings.NewReader(tc.Before))
		if err != nil {
			t.Error(err)
			continue
		}
		transform(gq.Find(`color`))
		buf := bytes.NewBuffer(nil)
		goquery.Render(buf, gq.Find(`body`).Children())
		if buf.String() != tc.After {
			t.Errorf("不相等：%s\nwant:\n%s\ngot:\n%s", tc.Before, tc.After, buf.String())
		}
	}
}

func TestPalette(t *testing.T) {
	fp, err := os.Create(`palette.html`)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	fp.WriteString(`<!DOCTYPE html><head><style>@media screen and (prefers-color-scheme: dark) { body { background-color: black; } }</style></head><body>`)

	var colors []struct {
		Name     string   `yaml:"name"`
		Hex      string   `yaml:"hex"`
		RGB      string   `yaml:"rgb"`
		Families []string `yaml:"families"`
	}

	y, err := os.Open(`colors.yaml`)
	if err != nil {
		panic(err)
	}
	defer y.Close()
	if err := yaml.NewDecoder(y).Decode(&colors); err != nil {
		panic(err)
	}

	for _, color := range colors {
		fmt.Fprintf(fp, `<div data-name="%s" style="display:inline-block;width:100px;height:100px;margin:1em;background-color:%s;">%s</div>`, color.Name, color.Hex, color.Name)
	}

	for _, color := range colors {
		fmt.Fprintf(fp, `<p data-name="%s" style="font-size: 200%%;color:%s;">%s</p>`, color.Name, color.Hex, color.Name)
	}

	fp.WriteString(`</body></html>`)
}

func TestAll(t *testing.T) {
	var testCases = []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `<p>a<color red>b</color>c</p>`,
			HTML:     `<p>a<span class="color fg-red">b</span>c</p>`,
		},
	}
	for i, tc := range testCases {
		md := renderers.NewMarkdown(New())
		html := utils.Must1(md.Render(tc.Markdown))
		if strings.TrimSpace(html) != tc.HTML {
			t.Fatalf("not equal: #%d\n%s\n\n%s", i, tc.HTML, html)
		}
	}
}
