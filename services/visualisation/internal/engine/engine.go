package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage2_layout"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage4_render"
	"github.com/google/uuid"
)

const (
	pdfPageMargin = 48.0
	pdfCoordScale = 16.0
	pdfLayerStep  = 140.0

	pdfSingleNodeWidth = 180.0
	pdfPairNodeWidth   = 300.0
	pdfNodeHeight      = 76.0
)

var errRootNotFound = errors.New("root person not found in tree content")

type relationView struct {
	from uuid.UUID
	to   uuid.UUID
	rt   familytreepb.RelationshipType
}

type personView struct {
	id     uuid.UUID
	label  string
	gender stage1_input.Gender
}

func RenderPDF(visType models.VisualisationType, rootPersonID uuid.UUID, includedPersonIDs []uuid.UUID, content *familytreepb.GetTreeContentResponse) ([]byte, error) {
	return RenderSVG(visType, rootPersonID, includedPersonIDs, content)
}

func RenderPDFWithTrace(visType models.VisualisationType, rootPersonID uuid.UUID, includedPersonIDs []uuid.UUID, content *familytreepb.GetTreeContentResponse, out io.Writer) ([]byte, error) {
	return RenderSVGWithTrace(visType, rootPersonID, includedPersonIDs, content, out)
}

func RenderSVG(visType models.VisualisationType, rootPersonID uuid.UUID, includedPersonIDs []uuid.UUID, content *familytreepb.GetTreeContentResponse) ([]byte, error) {
	return renderSVGInternal(visType, rootPersonID, includedPersonIDs, content, nil)
}

func RenderSVGWithTrace(visType models.VisualisationType, rootPersonID uuid.UUID, includedPersonIDs []uuid.UUID, content *familytreepb.GetTreeContentResponse, out io.Writer) ([]byte, error) {
	return renderSVGInternal(visType, rootPersonID, includedPersonIDs, content, out)
}

func renderSVGInternal(visType models.VisualisationType, rootPersonID uuid.UUID, includedPersonIDs []uuid.UUID, content *familytreepb.GetTreeContentResponse, out io.Writer) ([]byte, error) {
	if content == nil {
		return nil, fmt.Errorf("visualisation content is empty")
	}
	debugf(out, "=== ENGINE TRACE ===")
	debugf(out, "input: vis_type=%s root=%s included=%d", visType, rootPersonID.String(), len(includedPersonIDs))

	people, relations, err := normalizeInput(content)
	if err != nil {
		return nil, err
	}
	debugf(out, "[stage1_input] normalized people=%d relations=%d", len(people), len(relations))
	if out != nil {
		debugPrintPeople(out, people)
		debugPrintRelations(out, relations)
	}
	if _, ok := people[rootPersonID]; !ok {
		return nil, errRootNotFound
	}

	allowed := make(map[uuid.UUID]struct{}, len(includedPersonIDs)+1)
	for _, id := range includedPersonIDs {
		allowed[id] = struct{}{}
	}
	allowed[rootPersonID] = struct{}{}

	filteredPeople, filteredRelations := filterByVisualisation(visType, rootPersonID, allowed, people, relations)
	debugf(out, "[filter] people=%d relations=%d", len(filteredPeople), len(filteredRelations))
	if out != nil {
		debugPrintPeople(out, filteredPeople)
		debugPrintRelations(out, filteredRelations)
	}
	if _, ok := filteredPeople[rootPersonID]; !ok {
		return nil, errRootNotFound
	}

	tree, idToInt, err := buildStageTree(filteredPeople, filteredRelations)
	if err != nil {
		return nil, err
	}
	debugf(out, "[stage1_input->tree] tree people=%d", len(tree.People))
	if out != nil {
		debugPrintTree(out, tree)
	}
	rootIntID, ok := idToInt[rootPersonID]
	if !ok {
		return nil, errRootNotFound
	}

	if visType == models.VisualisationTypeDescendants {
		dropParents(tree)
	}
	if visType == models.VisualisationTypeAncestors {
		dropChildren(tree)
	}
	if out != nil {
		debugf(out, "[type-cut] after type-specific pruning")
		debugPrintTree(out, tree)
	}

	history, err := stage2_layout.LayoutFromPerson(tree, rootIntID)
	if err != nil {
		return nil, err
	}
	debugf(out, "[stage2_layout] records=%d", len(history.Records))
	if out != nil {
		debugPrintLayouts(out, tree)
		debugPrintHistory(out, history)
	}

	layouts := make(map[int]*stage1_input.PersonLayout)
	for id, person := range tree.People {
		if person.Layout != nil {
			layouts[id] = person.Layout
		}
	}

	start := tree.People[rootIntID]
	if start == nil || start.Layout == nil {
		return nil, errRootNotFound
	}

	om := stage3_ordering.ProcessPlacementHistory(history, start, start.Layout.Layer, layouts)
	if out != nil {
		debugPrintOrder(out, om)
	}
	cm := om.BuildCoordMatrix()
	if out != nil {
		debugPrintCoordMatrix(out, cm)
	}
	renderResult := stage4_render.BuildCoordRenderResult(cm, tree)
	debugf(out, "[stage4_render] nodes=%d edges=%d", len(renderResult.Nodes), len(renderResult.Edges))

	return renderLegacyLayoutToSVG(renderResult, tree), nil
}

func renderLegacyLayoutToSVG(result *stage4_render.CoordRenderResult, tree *stage1_input.FamilyTree) []byte {
	if result == nil {
		return []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><svg xmlns=\"http://www.w3.org/2000/svg\" width=\"1\" height=\"1\"></svg>")
	}
	renderer := stage4_render.NewCoordSVGRenderer(result, tree)
	return []byte(renderer.Render())
}

func debugf(out io.Writer, format string, args ...any) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, format+"\n", args...)
}

func debugPrintPeople(out io.Writer, people map[uuid.UUID]personView) {
	ids := make([]uuid.UUID, 0, len(people))
	for id := range people {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	for _, id := range ids {
		p := people[id]
		debugf(out, "  person: %s | %s", id.String(), p.label)
	}
}

func debugPrintRelations(out io.Writer, relations []relationView) {
	for _, rel := range relations {
		debugf(out, "  rel: %s -> %s (%s)", rel.from.String(), rel.to.String(), rel.rt.String())
	}
}

func debugPrintTree(out io.Writer, tree *stage1_input.FamilyTree) {
	ids := make([]int, 0, len(tree.People))
	for id := range tree.People {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		p := tree.People[id]
		mother := "-"
		father := "-"
		if p.Mother != nil {
			mother = p.Mother.Name
		}
		if p.Father != nil {
			father = p.Father.Name
		}
		debugf(out, "  tree[%d]=%s | mother=%s father=%s partners=%d children=%d", p.ID, p.Name, mother, father, len(p.Partners), len(p.Children))
	}
}

func debugPrintLayouts(out io.Writer, tree *stage1_input.FamilyTree) {
	ids := make([]int, 0, len(tree.People))
	for id := range tree.People {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		p := tree.People[id]
		if p.Layout == nil {
			debugf(out, "  layout[%d]=<nil>", p.ID)
			continue
		}
		debugf(out, "  layout[%d]=layer:%d processed:%t", p.ID, p.Layout.Layer, p.Layout.Processed)
	}
}

func debugPrintHistory(out io.Writer, history *stage1_input.PlacementHistory) {
	for i, record := range history.Records {
		if record.AddedPerson2 != nil {
			debugf(out, "  history[%d]: %s -> [%s,%s] type=%s dir=%s", i, record.FromPerson.Name, record.AddedPerson.Name, record.AddedPerson2.Name, record.RelationType.String(), record.Direction.String())
			continue
		}
		debugf(out, "  history[%d]: %s -> %s type=%s dir=%s", i, record.FromPerson.Name, record.AddedPerson.Name, record.RelationType.String(), record.Direction.String())
	}
}

func debugPrintOrder(out io.Writer, om *stage3_ordering.OrderManager) {
	debugf(out, "[stage3_ordering] layers order")
	for _, layer := range om.GetAllLayers() {
		debugf(out, "  layer %d: people=%v nodes=%d", layer.Number, layer.GetPeopleIDs(), len(layer.GetNodes()))
	}
}

func debugPrintCoordMatrix(out io.Writer, cm *stage3_ordering.CoordMatrix) {
	debugf(out, "[stage3_ordering] coord matrix layers=%d..%d", cm.MinLayer, cm.MaxLayer)
	for layerNum := cm.MaxLayer; layerNum >= cm.MinLayer; layerNum-- {
		nodes := cm.Layers[layerNum]
		for _, node := range nodes {
			if node.IsPseudo {
				debugf(out, "  cm[%d]: pseudo left=%d right=%d", layerNum, node.Left, node.Right)
				continue
			}
			names := make([]string, 0, len(node.People))
			for _, person := range node.People {
				names = append(names, person.Name)
			}
			debugf(out, "  cm[%d]: %v left=%d right=%d", layerNum, names, node.Left, node.Right)
		}
	}
}

func normalizeInput(content *familytreepb.GetTreeContentResponse) (map[uuid.UUID]personView, []relationView, error) {
	people := make(map[uuid.UUID]personView, len(content.GetPersons()))
	for _, person := range content.GetPersons() {
		id, err := uuid.Parse(person.GetId())
		if err != nil {
			continue
		}
		people[id] = personView{
			id:     id,
			label:  buildLabel(person),
			gender: mapGender(person.GetGender()),
		}
	}

	relations := make([]relationView, 0, len(content.GetRelationships()))
	for _, rel := range content.GetRelationships() {
		fromID, err := uuid.Parse(rel.GetPersonIdFrom())
		if err != nil {
			continue
		}
		toID, err := uuid.Parse(rel.GetPersonIdTo())
		if err != nil {
			continue
		}
		if _, ok := people[fromID]; !ok {
			continue
		}
		if _, ok := people[toID]; !ok {
			continue
		}
		relations = append(relations, relationView{from: fromID, to: toID, rt: rel.GetType()})
	}

	if len(people) == 0 {
		return nil, nil, fmt.Errorf("visualisation content does not contain valid people")
	}

	return people, relations, nil
}

func filterByVisualisation(
	visType models.VisualisationType,
	rootID uuid.UUID,
	allowed map[uuid.UUID]struct{},
	people map[uuid.UUID]personView,
	relations []relationView,
) (map[uuid.UUID]personView, []relationView) {
	if visType == models.VisualisationTypeFull {
		filteredPeople := make(map[uuid.UUID]personView, len(allowed))
		for id := range allowed {
			if p, ok := people[id]; ok {
				filteredPeople[id] = p
			}
		}
		filteredRelations := make([]relationView, 0, len(relations))
		for _, rel := range relations {
			if _, ok := filteredPeople[rel.from]; !ok {
				continue
			}
			if _, ok := filteredPeople[rel.to]; !ok {
				continue
			}
			filteredRelations = append(filteredRelations, rel)
		}
		return filteredPeople, filteredRelations
	}

	filteredPeople := make(map[uuid.UUID]personView, len(people))
	for id, p := range people {
		filteredPeople[id] = p
	}
	filteredRelations := append([]relationView(nil), relations...)

	if visType == models.VisualisationTypeAncestors {
		keep := ancestorsReachable(rootID, filteredRelations)
		for id := range filteredPeople {
			if _, ok := keep[id]; !ok {
				delete(filteredPeople, id)
			}
		}
		filteredRelations = keepRelationsWithin(filteredRelations, keep)
	}

	if visType == models.VisualisationTypeDescendants {
		keep := descendantsReachable(rootID, filteredRelations)
		for id := range filteredPeople {
			if _, ok := keep[id]; !ok {
				delete(filteredPeople, id)
			}
		}
		filteredRelations = keepRelationsWithin(filteredRelations, keep)
	}

	return filteredPeople, filteredRelations
}

func ancestorsReachable(root uuid.UUID, relations []relationView) map[uuid.UUID]struct{} {
	parentsByChild := make(map[uuid.UUID][]uuid.UUID)
	for _, rel := range relations {
		if rel.rt == familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD {
			parentsByChild[rel.to] = appendUniqueUUID(parentsByChild[rel.to], rel.from)
		}
	}
	visited := map[uuid.UUID]struct{}{root: {}}
	queue := []uuid.UUID{root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, parent := range parentsByChild[current] {
			if _, ok := visited[parent]; ok {
				continue
			}
			visited[parent] = struct{}{}
			queue = append(queue, parent)
		}
	}
	return visited
}

func descendantsReachable(root uuid.UUID, relations []relationView) map[uuid.UUID]struct{} {
	childrenByParent := make(map[uuid.UUID][]uuid.UUID)
	for _, rel := range relations {
		if rel.rt == familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD {
			childrenByParent[rel.from] = appendUniqueUUID(childrenByParent[rel.from], rel.to)
		}
	}
	visited := map[uuid.UUID]struct{}{root: {}}
	queue := []uuid.UUID{root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, child := range childrenByParent[current] {
			if _, ok := visited[child]; ok {
				continue
			}
			visited[child] = struct{}{}
			queue = append(queue, child)
		}
	}
	return visited
}

func keepRelationsWithin(relations []relationView, keep map[uuid.UUID]struct{}) []relationView {
	result := make([]relationView, 0, len(relations))
	for _, rel := range relations {
		if _, ok := keep[rel.from]; !ok {
			continue
		}
		if _, ok := keep[rel.to]; !ok {
			continue
		}
		result = append(result, rel)
	}
	return result
}

func buildStageTree(people map[uuid.UUID]personView, relations []relationView) (*stage1_input.FamilyTree, map[uuid.UUID]int, error) {
	ids := make([]uuid.UUID, 0, len(people))
	for id := range people {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})

	tree := stage1_input.NewFamilyTree()
	idToInt := make(map[uuid.UUID]int, len(ids))
	for i, id := range ids {
		pid := i + 1
		idToInt[id] = pid
		p := people[id]
		tree.AddPerson(stage1_input.NewPerson(pid, p.label, p.gender))
	}

	parentsByChild := make(map[uuid.UUID][]uuid.UUID)
	partnerType := make(map[[2]uuid.UUID]familytreepb.RelationshipType)
	for _, rel := range relations {
		switch rel.rt {
		case familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD:
			parentsByChild[rel.to] = appendUniqueUUID(parentsByChild[rel.to], rel.from)
		case familytreepb.RelationshipType_RELATIONSHIP_PARTNER,
			familytreepb.RelationshipType_RELATIONSHIP_PARTNER_UNMARRIED,
			familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED,
			familytreepb.RelationshipType_RELATIONSHIP_PARTNER_DIVORCED:
			key := pairKey(rel.from, rel.to)
			if old, ok := partnerType[key]; !ok || partnerPriority(rel.rt) < partnerPriority(old) {
				partnerType[key] = rel.rt
			}
		}
	}

	for childID, parentIDs := range parentsByChild {
		motherID, fatherID := pickParents(parentIDs, people)
		if motherID == uuid.Nil || fatherID == uuid.Nil {
			continue
		}
		childInt := idToInt[childID]
		motherInt := idToInt[motherID]
		fatherInt := idToInt[fatherID]
		if err := tree.SetParents(childInt, motherInt, fatherInt); err != nil {
			return nil, nil, err
		}
	}

	type partnerEdge struct {
		left  uuid.UUID
		right uuid.UUID
		rt    familytreepb.RelationshipType
	}
	edges := make([]partnerEdge, 0, len(partnerType))
	for key, rt := range partnerType {
		edges = append(edges, partnerEdge{left: key[0], right: key[1], rt: rt})
	}
	sort.SliceStable(edges, func(i, j int) bool {
		pi := partnerPriority(edges[i].rt)
		pj := partnerPriority(edges[j].rt)
		if pi != pj {
			return pi < pj
		}
		if edges[i].left.String() != edges[j].left.String() {
			return edges[i].left.String() < edges[j].left.String()
		}
		return edges[i].right.String() < edges[j].right.String()
	})

	for _, edge := range edges {
		left := idToInt[edge.left]
		right := idToInt[edge.right]
		if err := tree.AddPartnership(left, right); err != nil {
			return nil, nil, err
		}
	}

	return tree, idToInt, nil
}

func dropParents(tree *stage1_input.FamilyTree) {
	for _, person := range tree.People {
		person.Mother = nil
		person.Father = nil
	}
}

func dropChildren(tree *stage1_input.FamilyTree) {
	for _, person := range tree.People {
		person.Children = nil
	}
}

func appendUniqueUUID(slice []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, existing := range slice {
		if existing == id {
			return slice
		}
	}
	return append(slice, id)
}

func pairKey(a, b uuid.UUID) [2]uuid.UUID {
	if a.String() < b.String() {
		return [2]uuid.UUID{a, b}
	}
	return [2]uuid.UUID{b, a}
}

func pickParents(parentIDs []uuid.UUID, people map[uuid.UUID]personView) (mother uuid.UUID, father uuid.UUID) {
	for _, id := range parentIDs {
		if people[id].gender == stage1_input.Female && mother == uuid.Nil {
			mother = id
		}
		if people[id].gender == stage1_input.Male && father == uuid.Nil {
			father = id
		}
	}
	for _, id := range parentIDs {
		if father == uuid.Nil && id != mother {
			father = id
			continue
		}
		if mother == uuid.Nil && id != father {
			mother = id
		}
	}
	return mother, father
}

func partnerPriority(rt familytreepb.RelationshipType) int {
	switch rt {
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED:
		return 0
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER:
		return 1
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER_UNMARRIED:
		return 2
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER_DIVORCED:
		return 3
	default:
		return 4
	}
}

func mapGender(g familytreepb.Gender) stage1_input.Gender {
	switch g {
	case familytreepb.Gender_GENDER_FEMALE:
		return stage1_input.Female
	default:
		return stage1_input.Male
	}
}

func buildLabel(person *familytreepb.Person) string {
	parts := make([]string, 0, 3)
	if first := strings.TrimSpace(person.GetFirstName()); first != "" {
		parts = append(parts, first)
	}
	if last := strings.TrimSpace(person.GetLastName()); last != "" {
		parts = append(parts, last)
	}
	if patronymic := strings.TrimSpace(person.GetPatronymic()); patronymic != "" {
		parts = append(parts, patronymic)
	}
	if len(parts) == 0 {
		return person.GetId()
	}
	return strings.Join(parts, " ")
}

func renderLegacyLayoutToPDF(result *stage4_render.CoordRenderResult, rootIntID int) []byte {
	if result == nil {
		return buildPDF(420, 320, nil)
	}

	nodeX := make([]float64, len(result.Nodes))
	nodeY := make([]float64, len(result.Nodes))
	nodeW := make([]float64, len(result.Nodes))
	maxRight := 0.0
	for i, node := range result.Nodes {
		x := pdfPageMargin + float64(node.Left)*pdfCoordScale
		w := pdfSingleNodeWidth
		if len(node.People) == 2 {
			w = pdfPairNodeWidth
		}
		y := pdfPageMargin + float64(result.MaxLayer-node.Layer)*pdfLayerStep
		nodeX[i] = x
		nodeY[i] = y
		nodeW[i] = w
		if x+w > maxRight {
			maxRight = x + w
		}
	}

	pageWidth := maxRight + pdfPageMargin
	if pageWidth < 640 {
		pageWidth = 640
	}
	layerCount := float64(result.MaxLayer - result.MinLayer + 1)
	if layerCount < 1 {
		layerCount = 1
	}
	pageHeight := pdfPageMargin*2 + layerCount*pdfLayerStep + pdfNodeHeight
	if pageHeight < 420 {
		pageHeight = 420
	}

	var content bytes.Buffer
	content.WriteString("0 0 0 RG\n")

	for _, edge := range result.Edges {
		switch edge.EdgeType {
		case "partner":
			content.WriteString("0.35 0.35 0.35 RG\n")
			y := pdfPageMargin + float64(result.MaxLayer-edge.FromY)*pdfLayerStep + pdfNodeHeight/2
			x1 := pdfPageMargin + float64(edge.FromX)*pdfCoordScale
			x2 := pdfPageMargin + float64(edge.ToX)*pdfCoordScale
			content.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n", x1, y, x2, y))
		case "parent-child":
			content.WriteString("0.25 0.25 0.25 RG\n")
			fromY := pdfPageMargin + float64(result.MaxLayer-edge.FromY)*pdfLayerStep + pdfNodeHeight
			toY := pdfPageMargin + float64(result.MaxLayer-edge.ToY)*pdfLayerStep
			midY := fromY + (toY-fromY)/2

			fromX := pdfPageMargin + float64(edge.FromX)*pdfCoordScale
			toX := pdfPageMargin + float64(edge.ToX)*pdfCoordScale
			if edge.ParentsAdjacent {
				fromX = pdfPageMargin + float64(edge.AdjacentCenterX)*pdfCoordScale
				content.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l %.2f %.2f l %.2f %.2f l S\n", fromX, fromY, fromX, midY, toX, midY, toX, toY))
			} else {
				offset := 1.0
				if edge.ParentAddedLeft {
					offset = fromX + pdfCoordScale
				} else {
					offset = fromX - pdfCoordScale
				}
				content.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l %.2f %.2f l %.2f %.2f l %.2f %.2f l S\n", fromX, fromY, offset, fromY, offset, midY, toX, midY, toX, toY))
			}
		}
	}

	for idx, node := range result.Nodes {
		x := nodeX[idx]
		y := nodeY[idx]
		w := nodeW[idx]

		fillR, fillG, fillB := 0.94, 0.94, 0.96
		if containsPerson(node.People, rootIntID) {
			fillR, fillG, fillB = 0.85, 0.92, 1.0
		} else if len(node.People) > 0 {
			if node.People[0].Gender == stage1_input.Female {
				fillR, fillG, fillB = 1.0, 0.90, 0.93
			} else {
				fillR, fillG, fillB = 0.86, 0.92, 1.0
			}
		}

		content.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg\n", fillR, fillG, fillB))
		content.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re f\n", x, y, w, pdfNodeHeight))
		content.WriteString("0 0 0 RG\n")
		content.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re S\n", x, y, w, pdfNodeHeight))

		if len(node.People) == 1 {
			content.WriteString(drawCenteredText(x, y, w, pdfNodeHeight, node.People[0].Name))
		} else if len(node.People) == 2 {
			half := w / 2
			content.WriteString(drawCenteredText(x, y, half, pdfNodeHeight, node.People[0].Name))
			content.WriteString(drawCenteredText(x+half, y, half, pdfNodeHeight, node.People[1].Name))
		}
	}

	return buildPDF(pageWidth, pageHeight, content.Bytes())
}

func containsPerson(people []*stage1_input.Person, id int) bool {
	for _, person := range people {
		if person.ID == id {
			return true
		}
	}
	return false
}

func drawCenteredText(x, y, width, height float64, text string) string {
	text = sanitizePDFText(text)
	lines := splitText(text, 24)
	if len(lines) == 0 {
		return ""
	}

	fontSize := 10.0
	lineHeight := 12.0
	blockHeight := float64(len(lines)-1) * lineHeight
	startY := y + height/2 + blockHeight/2 + fontSize/2 - 2

	var out strings.Builder
	for i, line := range lines {
		textWidth := estimateTextWidth(line, fontSize)
		textX := x + (width-textWidth)/2
		textY := startY - float64(i)*lineHeight
		out.WriteString(fmt.Sprintf("BT /F1 %.1f Tf %.2f %.2f Td (%s) Tj ET\n", fontSize, textX, textY, escapePDF(line)))
	}
	return out.String()
}

func splitText(text string, maxLen int) []string {
	if text == "" {
		return nil
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return []string{text}
	}

	lines := make([]string, 0)
	current := parts[0]
	for _, part := range parts[1:] {
		if len(current)+1+len(part) <= maxLen {
			current += " " + part
			continue
		}
		lines = append(lines, current)
		current = part
	}
	lines = append(lines, current)
	if len(lines) > 3 {
		return lines[:3]
	}
	return lines
}

func estimateTextWidth(text string, fontSize float64) float64 {
	return float64(len(text)) * fontSize * 0.55
}

func sanitizePDFText(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return text
}

func escapePDF(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "(", "\\(")
	text = strings.ReplaceAll(text, ")", "\\)")
	return text
}

func buildPDF(width, height float64, content []byte) []byte {
	objects := make([][]byte, 0, 5)
	objects = append(objects, []byte("<< /Type /Catalog /Pages 2 0 R >>"))
	objects = append(objects, []byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"))
	objects = append(objects, []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", width, height)))
	objects = append(objects, []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"))
	stream := append([]byte(fmt.Sprintf("<< /Length %d >>\nstream\n", len(content))), content...)
	stream = append(stream, []byte("endstream")...)
	objects = append(objects, stream)

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n%EOF\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		buf.WriteString(fmt.Sprintf("%d 0 obj\n", i+1))
		buf.Write(obj)
		buf.WriteString("\nendobj\n")
	}
	xrefStart := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefStart))
	return buf.Bytes()
}
