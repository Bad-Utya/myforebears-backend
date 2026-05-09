package stage4_render

import (
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
	"html"
	"os"
)

const (
	singleNodeWidth = 170
	pairNodeWidth   = 280
	nodeHeight      = 80
	nodeSpacingY    = 80
	nodeSpacingX    = 30
	padding         = 50

	coordScale       = 85
	maleColor        = "#a8d5ff"
	femaleColor      = "#ffb6c1"
	unknownColor     = "#e0e0e0"
	strokeColor      = "#333333"
	textColor        = "#000000"
	partnerEdgeColor = "#494949"
	childEdgeColor   = "#494949"
)

type CoordSVGRenderer struct {
	result      *CoordRenderResult
	tree        *stage1_input.FamilyTree
	nodePixelX  map[int]int
	nodeCenterX map[int]int
	svgWidth    int
	svgHeight   int
}

func NewCoordSVGRenderer(result *CoordRenderResult, tree *stage1_input.FamilyTree) *CoordSVGRenderer {
	r := &CoordSVGRenderer{
		result:      result,
		tree:        tree,
		nodePixelX:  make(map[int]int),
		nodeCenterX: make(map[int]int),
	}
	r.calculateLayout()
	return r
}

func (r *CoordSVGRenderer) calculateLayout() {

	r.svgWidth = 0

	for i, node := range r.result.Nodes {

		x := padding + node.Left*coordScale

		width := singleNodeWidth
		if len(node.People) == 2 {
			width = pairNodeWidth
		}

		r.nodePixelX[i] = x
		r.nodeCenterX[i] = x + width/2

		rightEdge := x + width + padding
		if rightEdge > r.svgWidth {
			r.svgWidth = rightEdge
		}
	}

	r.svgHeight = padding*2 + (r.result.MaxLayer-r.result.MinLayer+1)*(nodeHeight+nodeSpacingY)
}

func (r *CoordSVGRenderer) getNodeColor(person *stage1_input.Person) string {
	switch person.Gender {
	case stage1_input.Male:
		return maleColor
	case stage1_input.Female:
		return femaleColor
	default:
		return unknownColor
	}
}

func (r *CoordSVGRenderer) coordToPixelX(coord int) int {
	return padding + coord*coordScale
}

func (r *CoordSVGRenderer) layerToPixelY(layer int) int {
	return padding + (r.result.MaxLayer-layer)*(nodeHeight+nodeSpacingY)
}

func (r *CoordSVGRenderer) Render() string {
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<style>
  .node { stroke: %s; stroke-width: 2; }
  .node-text { font-family: Arial, sans-serif; font-size: 12px; fill: %s; text-anchor: middle; }
  .edge-partner { stroke: %s; stroke-width: 2; fill: none; }
  .edge-child { stroke: %s; stroke-width: 2; fill: none; }
</style>
<rect width="100%%" height="100%%" fill="white"/>
`, r.svgWidth, r.svgHeight, r.svgWidth, r.svgHeight, strokeColor, textColor, partnerEdgeColor, childEdgeColor)

	svg += r.renderEdges()

	svg += r.renderNodes()

	svg += "</svg>"
	return svg
}

func (r *CoordSVGRenderer) renderEdges() string {
	result := ""

	for _, edge := range r.result.Edges {
		if edge.EdgeType == "partner" {

			if edge.FromNodeIdx == edge.ToNodeIdx {

				nodeIdx := edge.FromNodeIdx
				x := r.nodePixelX[nodeIdx]
				y := r.layerToPixelY(edge.FromY) + nodeHeight/2

				halfWidth := pairNodeWidth / 2
				x1 := x + halfWidth/2
				x2 := x + halfWidth + halfWidth/2

				result += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="edge-partner"/>
`, x1, y, x2, y)
			} else {

				x1 := r.coordToPixelX(edge.FromX)
				x2 := r.coordToPixelX(edge.ToX)
				y := r.layerToPixelY(edge.FromY) + nodeHeight/2

				result += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="edge-partner"/>
`, x1, y, x2, y)
			}
		} else {

			y1 := r.layerToPixelY(edge.FromY) + (nodeHeight / 2)
			y2 := r.layerToPixelY(edge.ToY)
			midY := y1 + (nodeHeight+nodeSpacingY)/2

			x2 := r.coordToPixelX(edge.ToX)

			if edge.ParentsAdjacent {

				startX := r.coordToPixelX(edge.AdjacentCenterX)

				result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d" class="edge-child"/>
`, startX, y1, startX, midY, x2, midY, x2, y2)
			} else {

				startX := r.coordToPixelX(edge.FromX)

				if edge.ParentAddedLeft {

					offsetX := r.coordToPixelX(edge.FromX + 1)
					result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d L %d %d" class="edge-child"/>
`, startX, y1, offsetX, y1, offsetX, midY, x2, midY, x2, y2)
				} else {

					offsetX := r.coordToPixelX(edge.FromX - 1)
					result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d L %d %d" class="edge-child"/>
`, startX, y1, offsetX, y1, offsetX, midY, x2, midY, x2, y2)
				}
			}
		}
	}

	return result
}

func (r *CoordSVGRenderer) renderNodes() string {
	result := ""

	for i, node := range r.result.Nodes {
		if len(node.People) == 0 {
			continue
		}

		x := r.nodePixelX[i]
		y := r.layerToPixelY(node.Layer)

		if len(node.People) == 1 {

			person := node.People[0]
			color := r.getNodeColor(person)

			rectWidth := int(float64(singleNodeWidth) * 0.8)
			rectX := x + int(float64(singleNodeWidth)*0.1)

			result += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="5" ry="5" fill="%s" class="node"/>
`, rectX, y, rectWidth, nodeHeight, color)

			line1, line2 := r.splitName(person.Name)
			line1 = r.truncateName(line1, 22)
			line2 = r.truncateName(line2, 22)

			textX := x + singleNodeWidth/2

			if line2 == "" {

				textY := y + nodeHeight/2 + 4
				result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY, html.EscapeString(line1))
			} else {

				textY1 := y + nodeHeight/2 - 8
				textY2 := y + nodeHeight/2 + 10
				result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY1, html.EscapeString(line1))
				result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY2, html.EscapeString(line2))
			}
		} else if len(node.People) == 2 {

			halfWidth := pairNodeWidth / 2

			for j, person := range node.People {
				personX := x + j*halfWidth
				color := r.getNodeColor(person)

				rectWidth := int(float64(halfWidth) * 0.6)
				rectX := personX + int(float64(halfWidth)*0.2)

				result += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="5" ry="5" fill="%s" class="node"/>
`, rectX, y, rectWidth, nodeHeight, color)

				line1, line2 := r.splitName(person.Name)
				line1 = r.truncateName(line1, 15)
				line2 = r.truncateName(line2, 15)

				textX := personX + halfWidth/2

				if line2 == "" {

					textY := y + nodeHeight/2 + 4
					result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY, html.EscapeString(line1))
				} else {

					textY1 := y + nodeHeight/2 - 8
					textY2 := y + nodeHeight/2 + 10
					result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY1, html.EscapeString(line1))
					result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY2, html.EscapeString(line2))
				}
			}
		}
	}

	return result
}

func (r *CoordSVGRenderer) splitName(name string) (string, string) {
	parts := []rune(name)
	var words []string
	currentWord := []rune{}

	for _, ch := range parts {
		if ch == ' ' {
			if len(currentWord) > 0 {
				words = append(words, string(currentWord))
				currentWord = []rune{}
			}
		} else {
			currentWord = append(currentWord, ch)
		}
	}
	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	if len(words) == 0 {
		return "", ""
	}
	if len(words) == 1 {
		return words[0], ""
	}
	return words[0], words[1]
}

func (r *CoordSVGRenderer) truncateName(name string, maxLen int) string {
	nameRunes := []rune(name)
	if len(nameRunes) > maxLen {
		return string(nameRunes[:maxLen-2]) + ".."
	}
	return name
}

func (r *CoordSVGRenderer) RenderToFile(filename string) error {
	svg := r.Render()
	return os.WriteFile(filename, []byte(svg), 0644)
}

func GenerateCoordSVG(cm *stage3_ordering.CoordMatrix, tree *stage1_input.FamilyTree, filename string) error {
	result := BuildCoordRenderResult(cm, tree)
	renderer := NewCoordSVGRenderer(result, tree)
	return renderer.RenderToFile(filename)
}
