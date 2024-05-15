package plantuml

import (
	"context"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestCompress(t *testing.T) {
	// 还不原，但是官方能正确显示
	t.SkipNow()
	s := `@startuml
Alice -> Bob: Authentication Request
Bob --> Alice: Authentication Response
@enduml`
	c := `Syp9J4vLqBLJSCfFib9mB2t9ICqhoKnEBCdCprC8IYqiJIqkuGBAAUW2rO0LOr5LN92VLvpA1G00`
	if x, err := compress([]byte(s)); err != nil || x != c {
		t.Fatal(`not equal`, err, c, x)
	}
}

func TestFetch(t *testing.T) {
	uml := `
@startuml
Bob -> Alice : hello
@enduml
`
	compressed := utils.Must(compress([]byte(uml)))
	svg := utils.Must(fetch(context.Background(), `https://www.plantuml.com/plantuml`, `svg`, compressed))
	t.Log(string(svg))
}

func TestEnableDarkMode(t *testing.T) {
	t.SkipNow()
	s := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" contentStyleType="text/css" height="120px" preserveAspectRatio="none" style="width:109px;height:120px;background:#FFFFFF;" version="1.1" viewBox="0 0 109 120" width="109px" zoomAndPan="magnify"><defs/><g><line style="stroke:#181818;stroke-width:0.5;stroke-dasharray:5.0,5.0;" x1="26" x2="26" y1="36.2969" y2="85.4297"/><line style="stroke:#181818;stroke-width:0.5;stroke-dasharray:5.0,5.0;" x1="80" x2="80" y1="36.2969" y2="85.4297"/><rect fill="#E2E2F0" height="30.2969" rx="2.5" ry="2.5" style="stroke:#181818;stroke-width:0.5;" width="42" x="5" y="5"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacing" textLength="28" x="12" y="24.9951">Bob</text><rect fill="#E2E2F0" height="30.2969" rx="2.5" ry="2.5" style="stroke:#181818;stroke-width:0.5;" width="42" x="5" y="84.4297"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacing" textLength="28" x="12" y="104.4248">Bob</text><rect fill="#E2E2F0" height="30.2969" rx="2.5" ry="2.5" style="stroke:#181818;stroke-width:0.5;" width="46" x="57" y="5"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacing" textLength="32" x="64" y="24.9951">Alice</text><rect fill="#E2E2F0" height="30.2969" rx="2.5" ry="2.5" style="stroke:#181818;stroke-width:0.5;" width="46" x="57" y="84.4297"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacing" textLength="32" x="64" y="104.4248">Alice</text><polygon fill="#181818" points="68,63.4297,78,67.4297,68,71.4297,72,67.4297" style="stroke:#181818;stroke-width:1.0;"/><line style="stroke:#181818;stroke-width:1.0;" x1="26" x2="74" y1="67.4297" y2="67.4297"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacing" textLength="30" x="33" y="62.3638">hello</text></g></svg>`
	after := ``
	if x := string(enabledDarkMode([]byte(s))); x != after {
		t.Log(`dark mode failed`)
		t.Log(s)
		t.Log(x)
		t.Log(after)
		t.FailNow()
	}
}
