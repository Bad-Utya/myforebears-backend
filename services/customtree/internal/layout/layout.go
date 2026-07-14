package layout

import (
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	"github.com/google/uuid"
	"html"
	"math"
	"math/rand"
	"sort"
	"strings"
)

type Node struct {
	EntityID      uuid.UUID
	Name          string
	AvatarPhotoID *uuid.UUID
	Layer         int
	X, Y          float64
}
type DrawEdge struct {
	ParentID, ChildID  uuid.UUID
	LabelDown, LabelUp string
}
type Result struct {
	Nodes         []Node
	Edges         []DrawEdge
	Width, Height float64
}

func Build(root uuid.UUID, entities []domain.Entity, edges []domain.Edge, down, up string) Result {
	byID := map[uuid.UUID]domain.Entity{}
	children := map[uuid.UUID][]uuid.UUID{}
	parent := map[uuid.UUID]uuid.UUID{}
	for _, e := range entities {
		byID[e.ID] = e
	}
	for _, e := range edges {
		children[e.ParentID] = append(children[e.ParentID], e.ChildID)
		parent[e.ChildID] = e.ParentID
	}
	layer := map[uuid.UUID]int{root: 0}
	queue := []uuid.UUID{root}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if p, ok := parent[id]; ok {
			if _, seen := layer[p]; !seen {
				layer[p] = layer[id] - 1
				queue = append(queue, p)
			}
		}
		for _, c := range children[id] {
			if _, seen := layer[c]; !seen {
				layer[c] = layer[id] + 1
				queue = append(queue, c)
			}
		}
	}
	layers := map[int][]uuid.UUID{}
	minL, maxL := 0, 0
	for id, l := range layer {
		layers[l] = append(layers[l], id)
		if l < minL {
			minL = l
		}
		if l > maxL {
			maxL = l
		}
	}
	for l := minL; l <= maxL; l++ {
		sort.Slice(layers[l], func(i, j int) bool { return layers[l][i].String() < layers[l][j].String() })
	}
	optimize(layers, edges, minL, maxL)
	const dx, dy = 220.0, 170.0
	maxCount := 1
	for _, ids := range layers {
		if len(ids) > maxCount {
			maxCount = len(ids)
		}
	}
	width := float64(maxCount) * dx
	nodes := []Node{}
	for l := minL; l <= maxL; l++ {
		ids := layers[l]
		offset := (float64(maxCount-len(ids)) * dx) / 2
		for i, id := range ids {
			e := byID[id]
			nodes = append(nodes, Node{EntityID: id, Name: e.Name, AvatarPhotoID: e.AvatarPhotoID, Layer: l, X: offset + float64(i)*dx + dx/2, Y: float64(l-minL)*dy + 80})
		}
	}
	outEdges := make([]DrawEdge, 0, len(edges))
	for _, e := range edges {
		outEdges = append(outEdges, DrawEdge{e.ParentID, e.ChildID, down, up})
	}
	return Result{nodes, outEdges, width, float64(maxL-minL+1)*dy + 80}
}
func optimize(layers map[int][]uuid.UUID, edges []domain.Edge, minL, maxL int) {
	rng := rand.New(rand.NewSource(42))
	cost := func() float64 {
		pos := map[uuid.UUID]int{}
		lay := map[uuid.UUID]int{}
		for l, ids := range layers {
			for i, id := range ids {
				pos[id] = i
				lay[id] = l
			}
		}
		c := 0.0
		for _, e := range edges {
			c += math.Abs(float64(pos[e.ParentID] - pos[e.ChildID]))
		}
		for i, a := range edges {
			for _, b := range edges[i+1:] {
				if lay[a.ParentID] == lay[b.ParentID] && lay[a.ChildID] == lay[b.ChildID] && (pos[a.ParentID]-pos[b.ParentID])*(pos[a.ChildID]-pos[b.ChildID]) < 0 {
					c += 20
				}
			}
		}
		return c
	}
	current := cost()
	temp := 10.0
	for i := 0; i < 2500; i++ {
		l := minL + rng.Intn(maxL-minL+1)
		ids := layers[l]
		if len(ids) < 2 {
			temp *= .997
			continue
		}
		a, b := rng.Intn(len(ids)), rng.Intn(len(ids))
		if a == b {
			continue
		}
		ids[a], ids[b] = ids[b], ids[a]
		next := cost()
		if next < current || rng.Float64() < math.Exp((current-next)/temp) {
			current = next
		} else {
			ids[a], ids[b] = ids[b], ids[a]
		}
		temp *= .997
		if temp < .05 {
			temp = .05
		}
	}
}
func (r Result) SVG() []byte {
	nodes := map[uuid.UUID]Node{}
	for _, n := range r.Nodes {
		nodes[n.EntityID] = n
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">`, r.Width, r.Height, r.Width, r.Height)
	b.WriteString(`<style>.n{fill:#fff;stroke:#334155;stroke-width:2}.t{font:14px sans-serif;fill:#0f172a}.e{stroke:#64748b;stroke-width:2;fill:none}.l{font:11px sans-serif;fill:#475569}</style>`)
	for _, e := range r.Edges {
		p, c := nodes[e.ParentID], nodes[e.ChildID]
		label := e.LabelDown + " / " + e.LabelUp
		fmt.Fprintf(&b, `<path class="e" d="M %.1f %.1f L %.1f %.1f"/><text class="l" x="%.1f" y="%.1f">%s</text>`, p.X, p.Y+30, c.X, c.Y-30, (p.X+c.X)/2, (p.Y+c.Y)/2, html.EscapeString(label))
	}
	for _, n := range r.Nodes {
		fmt.Fprintf(&b, `<rect class="n" x="%.1f" y="%.1f" width="180" height="60" rx="10"/><text class="t" x="%.1f" y="%.1f" text-anchor="middle">%s</text>`, n.X-90, n.Y-30, n.X, n.Y+5, html.EscapeString(n.Name))
	}
	b.WriteString(`</svg>`)
	return []byte(b.String())
}
