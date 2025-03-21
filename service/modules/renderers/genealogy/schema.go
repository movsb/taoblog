package genealogy

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	graphviz "github.com/goccy/go-graphviz"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
	"github.com/movsb/taoblog/service/modules/renderers/reminders/lunar"
)

type Date struct {
	Solar lunar.SolarDate
	Lunar lunar.LunarDate
}

func (d *Date) String() string {
	t := time.Time(d.Solar)
	if !t.IsZero() {
		return t.Format(time.DateOnly)
	}
	if !d.Lunar.IsZero() {
		return d.Lunar.DateString()
	}
	return ``
}

func (d *Date) UnmarshalYAML(unmarshal func(any) error) (outErr error) {
	defer utils.CatchAsError(&outErr)

	var raw string
	utils.Must(unmarshal(&raw))

	if strings.TrimSpace(raw) == `` {
		return nil
	}

	if dd, err := reminders.NewUserDateFromString(raw); err == nil {
		d.Solar = lunar.SolarDate(dd.Time)
		return nil
	}

	dates := strings.Split(raw, `,`)
	for _, date := range dates {
		if dd, err := reminders.NewUserDateFromString(date); err == nil {
			d.Solar = lunar.SolarDate(dd.Time)
			continue
		}
		if dd, err := lunar.ParseLunarDate(date); err == nil {
			d.Lunar = *dd
			continue
		}
		return fmt.Errorf(`无法解析日期：%s`, date)
	}

	if time.Time(d.Solar).IsZero() && d.Lunar.IsZero() {
		return fmt.Errorf(`无法解析日期：%s`, raw)
	}

	return nil
}

type Individual struct {
	// 唯一标识。
	// 如果为空，取名字。
	ID string `yaml:"id"`

	// 本人的名字。
	Name string `yaml:"name"`

	// 配偶。只需单方记录。
	Spouse string `yaml:"spouse"`

	// 直系亲属（上级）。
	Father string `yaml:"father"`
	Mother string `yaml:"mother"`

	// 生日 & 逝世
	Birth Date `yaml:"birth"`
	Death Date `yaml:"death"`
}

// 从测试代码偷过来的，写得很乱。
func gen(w io.Writer, individuals []*Individual) {
	g, _ := graphviz.New(context.Background())
	defer g.Close()
	graph, _ := g.Graph()
	graph.SetPad(.5)
	graph.SetRankDir(graphviz.TBRank)
	{
		var nodes []*graphviz.Node
		var maps = map[*graphviz.Node]*Individual{}
		var spouses = map[string]string{}
		for _, p := range individuals {
			node, _ := graph.CreateNodeByName(string(p.ID))
			node.SetLabel(p.Name)
			node.SetShape(graphviz.BoxShape)
			nodes = append(nodes, node)
			maps[node] = p
		}
		paired := func(id1, id2 string) bool {
			p1 := spouses[string(id1)]
			p2 := spouses[string(id2)]
			if p1 == string(id2) || p2 == string(id1) {
				return true
			}
			return false
		}
		for _, n1 := range nodes {
			for _, n2 := range nodes {
				if n1 == n2 {
					continue
				}
				if maps[n2].Father == maps[n1].ID {
					n1n, _ := n1.Name()
					n2n, _ := n2.Name()
					e, _ := graph.CreateEdgeByName(fmt.Sprintf(`%s->%s`, n2n, n1n), n1, n2)
					e.SetComment("father")
				}
				if maps[n2].Mother == maps[n1].ID {
					n1n, _ := n1.Name()
					n2n, _ := n2.Name()
					e, _ := graph.CreateEdgeByName(fmt.Sprintf(`%s->%s`, n2n, n1n), n1, n2)
					e.SetComment("mother")
				}
				if maps[n2].Spouse == maps[n1].ID && !paired(maps[n1].ID, maps[n2].ID) {
					spouses[maps[n1].ID] = maps[n2].ID
					n1n, _ := n1.Name()
					n2n, _ := n2.Name()

					// 创建子图以确保配偶水平对齐
					subgraph, _ := graph.CreateSubGraphByName(fmt.Sprintf("%s_%s", n1n, n2n))
					e, _ := subgraph.CreateEdgeByName(fmt.Sprintf(`%s->%s`, n2n, n1n), n1, n2)
					e.SetComment("spouse")
					e.SetArrowHead(graphviz.NoneArrow)
					e.SetArrowTail(graphviz.NoneArrow)
					e.SetConstraint(false)
				}
			}
		}
	}
	g.Render(context.Background(), graph, graphviz.SVG, w)
}
