package stage4_render

import (
	"fmt"
	"html"
	"os"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

const (
	// Р Р°Р·РјРµСЂС‹ СЌР»РµРјРµРЅС‚РѕРІ
	singleNodeWidth = 170 // С€РёСЂРёРЅР° РѕРґРёРЅРѕС‡РЅРѕР№ РІРµСЂС€РёРЅС‹
	pairNodeWidth   = 280 // С€РёСЂРёРЅР° РІРµСЂС€РёРЅС‹ СЃ РїР°СЂРѕР№
	nodeHeight      = 80  // РІС‹СЃРѕС‚Р° СѓРІРµР»РёС‡РµРЅР° РІ 2 СЂР°Р·Р° РґР»СЏ РґРІСѓС… СЃС‚СЂРѕРє С‚РµРєСЃС‚Р°
	nodeSpacingY    = 80
	nodeSpacingX    = 30 // РіРѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅС‹Р№ РѕС‚СЃС‚СѓРї РјРµР¶РґСѓ СѓР·Р»Р°РјРё
	padding         = 50

	// РњР°СЃС€С‚Р°Р± РєРѕРѕСЂРґРёРЅР°С‚ (1 РµРґРёРЅРёС†Р° РєРѕРѕСЂРґРёРЅР°С‚С‹ = СЃС‚РѕР»СЊРєРѕ РїРёРєСЃРµР»РµР№)
	coordScale       = 85
	maleColor        = "#a8d5ff"
	femaleColor      = "#ffb6c1"
	unknownColor     = "#e0e0e0"
	strokeColor      = "#333333"
	textColor        = "#000000"
	partnerEdgeColor = "#494949"
	childEdgeColor   = "#494949"
)

// CoordSVGRenderer РіРµРЅРµСЂРёСЂСѓРµС‚ SVG РёР· CoordMatrix
type CoordSVGRenderer struct {
	result      *CoordRenderResult
	tree        *stage1_input.FamilyTree
	nodePixelX  map[int]int // РЅРѕРјРµСЂ СѓР·Р»Р° -> X РІ РїРёРєСЃРµР»СЏС…
	nodeCenterX map[int]int // РЅРѕРјРµСЂ СѓР·Р»Р° -> С†РµРЅС‚СЂ X РІ РїРёРєСЃРµР»СЏС…
	svgWidth    int
	svgHeight   int
}

// NewCoordSVGRenderer СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ СЂРµРЅРґРµСЂРµСЂ
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

// calculateLayout СЂР°СЃСЃС‡РёС‚С‹РІР°РµС‚ РїРѕР·РёС†РёРё РІСЃРµС… СѓР·Р»РѕРІ РЅР° РѕСЃРЅРѕРІРµ РёС… РєРѕРѕСЂРґРёРЅР°С‚
func (r *CoordSVGRenderer) calculateLayout() {
	// Р”Р»СЏ РєР°Р¶РґРѕРіРѕ СѓР·Р»Р° СЂР°СЃСЃС‡РёС‚С‹РІР°РµРј РїРёРєСЃРµР»СЊРЅСѓСЋ РїРѕР·РёС†РёСЋ РЅР° РѕСЃРЅРѕРІРµ РєРѕРѕСЂРґРёРЅР°С‚
	// РСЃРїРѕР»СЊР·СѓРµРј РєРѕРѕСЂРґРёРЅР°С‚С‹ Left РєР°Рє РѕСЃРЅРѕРІСѓ РґР»СЏ РїРѕР·РёС†РёРѕРЅРёСЂРѕРІР°РЅРёСЏ
	r.svgWidth = 0

	for i, node := range r.result.Nodes {
		// X РїРѕР·РёС†РёСЏ РЅР° РѕСЃРЅРѕРІРµ РєРѕРѕСЂРґРёРЅР°С‚С‹ Left
		x := padding + node.Left*coordScale

		// РЁРёСЂРёРЅР° СѓР·Р»Р°
		width := singleNodeWidth
		if len(node.People) == 2 {
			width = pairNodeWidth
		}

		r.nodePixelX[i] = x
		r.nodeCenterX[i] = x + width/2

		// РћР±РЅРѕРІР»СЏРµРј РјР°РєСЃРёРјР°Р»СЊРЅСѓСЋ С€РёСЂРёРЅСѓ SVG
		rightEdge := x + width + padding
		if rightEdge > r.svgWidth {
			r.svgWidth = rightEdge
		}
	}

	// Р’С‹СЃРѕС‚Р° SVG
	r.svgHeight = padding*2 + (r.result.MaxLayer-r.result.MinLayer+1)*(nodeHeight+nodeSpacingY)
}

// getNodeColor РІРѕР·РІСЂР°С‰Р°РµС‚ С†РІРµС‚ СѓР·Р»Р° РІ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё РѕС‚ РїРѕР»Р°
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

// coordToPixelX РїСЂРµРѕР±СЂР°Р·СѓРµС‚ X-РєРѕРѕСЂРґРёРЅР°С‚Сѓ РІ РїРёРєСЃРµР»Рё
func (r *CoordSVGRenderer) coordToPixelX(coord int) int {
	return padding + coord*coordScale
}

// layerToPixelY РїСЂРµРѕР±СЂР°Р·СѓРµС‚ РЅРѕРјРµСЂ СЃР»РѕСЏ РІ РїРёРєСЃРµР»Рё (РёРЅРІРµСЂС‚РёСЂСѓРµРј Y)
func (r *CoordSVGRenderer) layerToPixelY(layer int) int {
	return padding + (r.result.MaxLayer-layer)*(nodeHeight+nodeSpacingY)
}

// Render РіРµРЅРµСЂРёСЂСѓРµС‚ SVG-СЃС‚СЂРѕРєСѓ
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

	// Р РёСЃСѓРµРј СЃРІСЏР·Рё (СЃРЅР°С‡Р°Р»Р°, С‡С‚РѕР±С‹ Р±С‹Р»Рё РїРѕРґ СѓР·Р»Р°РјРё)
	svg += r.renderEdges()

	// Р РёСЃСѓРµРј СѓР·Р»С‹
	svg += r.renderNodes()

	svg += "</svg>"
	return svg
}

// renderEdges СЂРёСЃСѓРµС‚ РІСЃРµ СЃРІСЏР·Рё
func (r *CoordSVGRenderer) renderEdges() string {
	result := ""

	for _, edge := range r.result.Edges {
		if edge.EdgeType == "partner" {
			// РџР°СЂС‚РЅС‘СЂС‹ вЂ” РєСЂР°СЃРЅР°СЏ РіРѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅР°СЏ Р»РёРЅРёСЏ
			if edge.FromNodeIdx == edge.ToNodeIdx {
				// Р’РЅСѓС‚СЂРё РѕРґРЅРѕР№ РІРµСЂС€РёРЅС‹ (СЃРєР»РµРµРЅРЅР°СЏ РїР°СЂР°)
				nodeIdx := edge.FromNodeIdx
				x := r.nodePixelX[nodeIdx]
				y := r.layerToPixelY(edge.FromY) + nodeHeight/2

				// Р›РёРЅРёСЏ РѕС‚ СЃРµСЂРµРґРёРЅС‹ РїРµСЂРІРѕР№ РїРѕР»РѕРІРёРЅС‹ РґРѕ СЃРµСЂРµРґРёРЅС‹ РІС‚РѕСЂРѕР№
				halfWidth := pairNodeWidth / 2
				x1 := x + halfWidth/2
				x2 := x + halfWidth + halfWidth/2

				result += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="edge-partner"/>
`, x1, y, x2, y)
			} else {
				// РњРµР¶РґСѓ РґРІСѓРјСЏ СЃРјРµР¶РЅС‹РјРё РІРµСЂС€РёРЅР°РјРё (MergePartner)
				x1 := r.coordToPixelX(edge.FromX)
				x2 := r.coordToPixelX(edge.ToX)
				y := r.layerToPixelY(edge.FromY) + nodeHeight/2

				result += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="edge-partner"/>
`, x1, y, x2, y)
			}
		} else {
			// Р РѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє вЂ” РІРµСЂС‚РёРєР°Р»СЊРЅР°СЏ СЃРІСЏР·СЊ СЃ РёР·РіРёР±РѕРј
			y1 := r.layerToPixelY(edge.FromY) + (nodeHeight / 2) // РЅРёР· СЂРѕРґРёС‚РµР»СЏ
			y2 := r.layerToPixelY(edge.ToY)                      // РІРµСЂС… СЂРµР±С‘РЅРєР°
			midY := y1 + (nodeHeight+nodeSpacingY)/2             // РїРѕР»СЃР»РѕСЏ РІРЅРёР·

			x2 := r.coordToPixelX(edge.ToX) // РєРѕРѕСЂРґРёРЅР°С‚Р° СЂРµР±С‘РЅРєР°

			if edge.ParentsAdjacent {
				// Р РѕРґРёС‚РµР»Рё вЂ” СЃРјРµР¶РЅС‹Рµ РІРµСЂС€РёРЅС‹
				// Р›РёРЅРёСЏ РЅР°С‡РёРЅР°РµС‚СЃСЏ РЅР° СЃР»РѕРµ СЂРѕРґРёС‚РµР»РµР№, X = С†РµРЅС‚СЂ РјРµР¶РґСѓ СЃРјРµР¶РЅС‹РјРё РІРµСЂС€РёРЅР°РјРё
				startX := r.coordToPixelX(edge.AdjacentCenterX)

				// РџСѓС‚СЊ: РѕС‚ С†РµРЅС‚СЂР° РјРµР¶РґСѓ СЂРѕРґРёС‚РµР»СЏРјРё РІРЅРёР· РЅР° РїРѕР»СЃР»РѕСЏ, Р·Р°С‚РµРј Рє СЂРµР±С‘РЅРєСѓ
				result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d" class="edge-child"/>
`, startX, y1, startX, midY, x2, midY, x2, y2)
			} else {
				// РћРґРёРЅРѕС‡РЅС‹Р№ СЂРѕРґРёС‚РµР»СЊ
				startX := r.coordToPixelX(edge.FromX)

				if edge.ParentAddedLeft {
					// Р РѕРґРёС‚РµР»СЊ Р±С‹Р» РґРѕР±Р°РІР»РµРЅ СЃР»РµРІР° вЂ” Р»РёРЅРёСЏ РёРґС‘С‚ РЅР° 1 РєРѕРѕСЂРґРёРЅР°С‚Сѓ РІРїСЂР°РІРѕ
					offsetX := r.coordToPixelX(edge.FromX + 1)
					result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d L %d %d" class="edge-child"/>
`, startX, y1, offsetX, y1, offsetX, midY, x2, midY, x2, y2)
				} else {
					// Р РѕРґРёС‚РµР»СЊ Р±С‹Р» РґРѕР±Р°РІР»РµРЅ СЃРїСЂР°РІР° вЂ” Р»РёРЅРёСЏ РёРґС‘С‚ РЅР° 1 РєРѕРѕСЂРґРёРЅР°С‚Сѓ РІР»РµРІРѕ
					offsetX := r.coordToPixelX(edge.FromX - 1)
					result += fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d L %d %d" class="edge-child"/>
`, startX, y1, offsetX, y1, offsetX, midY, x2, midY, x2, y2)
				}
			}
		}
	}

	return result
}

// renderNodes СЂРёСЃСѓРµС‚ РІСЃРµ СѓР·Р»С‹
func (r *CoordSVGRenderer) renderNodes() string {
	result := ""

	for i, node := range r.result.Nodes {
		if len(node.People) == 0 {
			continue
		}

		// РџРѕР·РёС†РёСЏ Рё СЂР°Р·РјРµСЂ РёР· РїСЂРµРґСЂР°СЃСЃС‡РёС‚Р°РЅРЅС‹С… РґР°РЅРЅС‹С…
		x := r.nodePixelX[i]
		y := r.layerToPixelY(node.Layer)

		if len(node.People) == 1 {
			// РћРґРёРЅРѕС‡РЅР°СЏ РІРµСЂС€РёРЅР°
			person := node.People[0]
			color := r.getNodeColor(person)

			// РџСЂСЏРјРѕСѓРіРѕР»СЊРЅРёРє Р·Р°РЅРёРјР°РµС‚ 60% С€РёСЂРёРЅС‹, СЃ РѕС‚СЃС‚СѓРїР°РјРё РїРѕ 20% СЃ РєР°Р¶РґРѕР№ СЃС‚РѕСЂРѕРЅС‹
			rectWidth := int(float64(singleNodeWidth) * 0.8)
			rectX := x + int(float64(singleNodeWidth)*0.1)

			result += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="5" ry="5" fill="%s" class="node"/>
`, rectX, y, rectWidth, nodeHeight, color)

			// Р Р°Р·РґРµР»СЏРµРј РёРјСЏ РЅР° РґРІРµ СЃС‚СЂРѕРєРё
			line1, line2 := r.splitName(person.Name)
			line1 = r.truncateName(line1, 22)
			line2 = r.truncateName(line2, 22)

			textX := x + singleNodeWidth/2

			if line2 == "" {
				// РћРґРЅР° СЃС‚СЂРѕРєР° - РїРѕ С†РµРЅС‚СЂСѓ
				textY := y + nodeHeight/2 + 4
				result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY, html.EscapeString(line1))
			} else {
				// Р”РІРµ СЃС‚СЂРѕРєРё
				textY1 := y + nodeHeight/2 - 8
				textY2 := y + nodeHeight/2 + 10
				result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY1, html.EscapeString(line1))
				result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY2, html.EscapeString(line2))
			}
		} else if len(node.People) == 2 {
			// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° вЂ” СЂРёСЃСѓРµРј РґРІРµ РїРѕР»РѕРІРёРЅРєРё
			halfWidth := pairNodeWidth / 2

			for j, person := range node.People {
				personX := x + j*halfWidth
				color := r.getNodeColor(person)

				// РџСЂСЏРјРѕСѓРіРѕР»СЊРЅРёРє Р·Р°РЅРёРјР°РµС‚ 60% С€РёСЂРёРЅС‹, СЃ РѕС‚СЃС‚СѓРїР°РјРё РїРѕ 20% СЃ РєР°Р¶РґРѕР№ СЃС‚РѕСЂРѕРЅС‹
				rectWidth := int(float64(halfWidth) * 0.6)
				rectX := personX + int(float64(halfWidth)*0.2)

				result += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="5" ry="5" fill="%s" class="node"/>
`, rectX, y, rectWidth, nodeHeight, color)

				// Р Р°Р·РґРµР»СЏРµРј РёРјСЏ РЅР° РґРІРµ СЃС‚СЂРѕРєРё
				line1, line2 := r.splitName(person.Name)
				line1 = r.truncateName(line1, 15)
				line2 = r.truncateName(line2, 15)

				textX := personX + halfWidth/2

				if line2 == "" {
					// РћРґРЅР° СЃС‚СЂРѕРєР° - РїРѕ С†РµРЅС‚СЂСѓ
					textY := y + nodeHeight/2 + 4
					result += fmt.Sprintf(`<text x="%d" y="%d" class="node-text">%s</text>
`, textX, textY, html.EscapeString(line1))
				} else {
					// Р”РІРµ СЃС‚СЂРѕРєРё
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

// splitName СЂР°Р·РґРµР»СЏРµС‚ РёРјСЏ РЅР° РґРІРµ СЃС‚СЂРѕРєРё РїРѕ РїСЂРѕР±РµР»Р°Рј
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

// truncateName РѕР±СЂРµР·Р°РµС‚ РёРјСЏ РґРѕ maxLen СЂСѓРЅ
func (r *CoordSVGRenderer) truncateName(name string, maxLen int) string {
	nameRunes := []rune(name)
	if len(nameRunes) > maxLen {
		return string(nameRunes[:maxLen-2]) + ".."
	}
	return name
}

// RenderToFile СЃРѕС…СЂР°РЅСЏРµС‚ SVG РІ С„Р°Р№Р»
func (r *CoordSVGRenderer) RenderToFile(filename string) error {
	svg := r.Render()
	return os.WriteFile(filename, []byte(svg), 0644)
}

// GenerateCoordSVG вЂ” С„СѓРЅРєС†РёСЏ РґР»СЏ РіРµРЅРµСЂР°С†РёРё SVG РёР· CoordMatrix
func GenerateCoordSVG(cm *stage3_ordering.CoordMatrix, tree *stage1_input.FamilyTree, filename string) error {
	result := BuildCoordRenderResult(cm, tree)
	renderer := NewCoordSVGRenderer(result, tree)
	return renderer.RenderToFile(filename)
}
