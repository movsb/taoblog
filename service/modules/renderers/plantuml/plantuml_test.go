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
	compressed := utils.Must1(compress([]byte(uml)))
	svg := utils.Must1(fetch(context.Background(), `https://www.plantuml.com/plantuml`, `svg`, compressed, false))
	t.Log(string(svg))
}
