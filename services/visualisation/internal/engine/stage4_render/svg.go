package stage4_render

import (
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

const (
	singleNodeWidth = 340
	pairNodeWidth   = 560
	nodeHeight      = 80
	nodeSpacingY    = 80
	nodeSpacingX    = 60
	padding         = 100

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
	personData  map[string]PersonRenderData
	nodePixelX  map[int]int
	nodeCenterX map[int]int
	svgWidth    int
	svgHeight   int
}

func NewCoordSVGRenderer(result *CoordRenderResult, tree *stage1_input.FamilyTree, personData map[string]PersonRenderData) *CoordSVGRenderer {
	r := &CoordSVGRenderer{
		result:      result,
		tree:        tree,
		personData:  personData,
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
	.node-text { font-family: Arial, sans-serif; font-size: 12px; fill: %s; text-anchor: start; }
	.node-placeholder { fill: #ffffff; opacity: 0.6; }
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

				if edge.RouteAbove && x2-x1 > coordScale {
					leftTurnX := x1 + coordScale/2
					rightTurnX := x2 - coordScale/2
					routeY := r.layerToPixelY(edge.RouteAboveLayer) - nodeSpacingY/2
					result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d L %d %d L %d %d" class="edge-partner"/>
`, x1, y, leftTurnX, y, leftTurnX, routeY, rightTurnX, routeY, rightTurnX, y, x2, y)
				} else {
					result += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="edge-partner"/>
`, x1, y, x2, y)
				}
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
			rectWidth := int(float64(singleNodeWidth) * 0.8)
			rectX := x + int(float64(singleNodeWidth)*0.1)
			result += r.renderPersonCard(rectX, y, rectWidth, nodeHeight, person)
		} else if len(node.People) == 2 {
			halfWidth := pairNodeWidth / 2

			for j, person := range node.People {
				personX := x + j*halfWidth
				rectWidth := int(float64(halfWidth) * 0.6)
				rectX := personX + int(float64(halfWidth)*0.2)
				result += r.renderPersonCard(rectX, y, rectWidth, nodeHeight, person)
			}
		}
	}

	return result
}

func (r *CoordSVGRenderer) renderPersonCard(x, y, width, height int, person *stage1_input.Person) string {
	if person == nil {
		return ""
	}

	color := r.getNodeColor(person)
	data := r.resolvePersonData(person)

	result := fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="5" ry="5" fill="%s" class="node"/>
`, x, y, width, height, color)

	imageAreaWidth := width / 3
	imagePadding := 4
	imageX := x + imagePadding
	imageY := y + imagePadding
	imageW := imageAreaWidth - imagePadding*2
	imageH := height - imagePadding*2
	if imageW < 1 {
		imageW = 1
	}
	if imageH < 1 {
		imageH = 1
	}

	if data.AvatarData != "" && data.AvatarMime != "" {
		result += fmt.Sprintf(`<image x="%d" y="%d" width="%d" height="%d" preserveAspectRatio="xMidYMid slice" href="data:%s;base64,%s"/>
`, imageX, imageY, imageW, imageH, data.AvatarMime, data.AvatarData)
	} else {
		result += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="3" ry="3" class="node-placeholder"/>
`, imageX, imageY, imageW, imageH)
	}

	textPaddingX := 6
	textAreaX := x + imageAreaWidth + textPaddingX
	textAreaWidth := width - imageAreaWidth - textPaddingX*2
	if textAreaWidth < 1 {
		textAreaWidth = 1
	}
	maxChars := r.estimateMaxChars(textAreaWidth)
	line1, line2 := r.wrapNameTwoLines(data.DisplayName, maxChars)
	line3 := r.truncateText(data.DateLine, maxChars)

	if line2 == "" {
		line1Y := y + 32
		line3Y := y + 58
		result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textAreaX, line1Y, html.EscapeString(line1))
		result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textAreaX, line3Y, html.EscapeString(line3))
		return result
	}

	line1Y := y + 24
	line2Y := y + 42
	line3Y := y + 60
	result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textAreaX, line1Y, html.EscapeString(line1))
	result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textAreaX, line2Y, html.EscapeString(line2))
	result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textAreaX, line3Y, html.EscapeString(line3))

	return result
}

func (r *CoordSVGRenderer) resolvePersonData(person *stage1_input.Person) PersonRenderData {
	if person == nil {
		return PersonRenderData{DisplayName: "", DateLine: "неиз"}
	}

	if r.personData != nil && person.ExternalID != "" {
		if data, ok := r.personData[person.ExternalID]; ok {
			if data.DisplayName == "" {
				data.DisplayName = r.fallbackPersonName(person)
			}
			if data.DateLine == "" {
				data.DateLine = "неиз"
			}
			return data
		}
	}

	return PersonRenderData{
		DisplayName: r.fallbackPersonName(person),
		DateLine:    "неиз",
	}
}

func (r *CoordSVGRenderer) fallbackPersonName(person *stage1_input.Person) string {
	if person == nil {
		return ""
	}
	if person.Name != "" {
		return person.Name
	}
	if person.ExternalID != "" {
		return person.ExternalID
	}
	return fmt.Sprintf("%d", person.ID)
}

func (r *CoordSVGRenderer) wrapNameTwoLines(name string, maxChars int) (string, string) {
	if maxChars <= 0 {
		return "", ""
	}
	words := strings.Fields(name)
	if len(words) == 0 {
		return "", ""
	}

	line1 := ""
	line2 := ""
	for _, word := range words {
		if line1 == "" {
			line1 = word
			continue
		}
		if r.runeLen(line1)+1+r.runeLen(word) <= maxChars {
			line1 = line1 + " " + word
			continue
		}

		if line2 == "" {
			line2 = word
			continue
		}
		if r.runeLen(line2)+1+r.runeLen(word) <= maxChars {
			line2 = line2 + " " + word
			continue
		}
		break
	}

	line1 = r.truncateText(line1, maxChars)
	line2 = r.truncateText(line2, maxChars)
	return line1, line2
}

func (r *CoordSVGRenderer) estimateMaxChars(width int) int {
	if width <= 0 {
		return 1
	}
	chars := width / 7
	if chars < 6 {
		return 6
	}
	return chars
}

func (r *CoordSVGRenderer) runeLen(value string) int {
	return len([]rune(value))
}

func (r *CoordSVGRenderer) truncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	textRunes := []rune(text)
	if len(textRunes) > maxLen {
		if maxLen <= 2 {
			return string(textRunes[:maxLen])
		}
		return string(textRunes[:maxLen-2]) + ".."
	}
	return text
}

func (r *CoordSVGRenderer) RenderToFile(filename string) error {
	svg := r.Render()
	return os.WriteFile(filename, []byte(svg), 0644)
}

func GenerateCoordSVG(cm *stage3_ordering.CoordMatrix, tree *stage1_input.FamilyTree, filename string) error {
	result := BuildCoordRenderResult(cm, tree)
	renderer := NewCoordSVGRenderer(result, tree, nil)
	return renderer.RenderToFile(filename)
}
