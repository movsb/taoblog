package genealogy

import (
	"context"
	"fmt"
	"io"

	graphviz "github.com/goccy/go-graphviz"
)

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
