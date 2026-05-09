package engine

import (
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
	person familytreepb.Person
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

	tree, idToInt, _, err := buildStageTree(filteredPeople, filteredRelations)
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

	return renderSVG(renderResult, tree), nil
}

func renderSVG(result *stage4_render.CoordRenderResult, tree *stage1_input.FamilyTree) []byte {
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
		personCopy := *person
		people[id] = personView{
			id:     id,
			label:  buildLabel(person),
			gender: mapGender(person.GetGender()),
			person: personCopy,
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

func buildStageTree(people map[uuid.UUID]personView, relations []relationView) (*stage1_input.FamilyTree, map[uuid.UUID]int, map[int]familytreepb.Person, error) {
	ids := make([]uuid.UUID, 0, len(people))
	for id := range people {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})

	tree := stage1_input.NewFamilyTree()
	idToInt := make(map[uuid.UUID]int, len(ids))
	internalPersons := make(map[int]familytreepb.Person, len(ids))
	for i, id := range ids {
		pid := i + 1
		idToInt[id] = pid
		p := people[id]
		tree.AddPerson(stage1_input.NewPerson(pid, p.label, p.gender))
		internalPersons[pid] = p.person
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
			return nil, nil, nil, err
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
			return nil, nil, nil, err
		}
	}

	return tree, idToInt, internalPersons, nil
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

func buildPersonDisplayName(firstName, lastName, patronymic, fallback string) string {
	parts := make([]string, 0, 3)
	if first := strings.TrimSpace(firstName); first != "" {
		parts = append(parts, first)
	}
	if last := strings.TrimSpace(lastName); last != "" {
		parts = append(parts, last)
	}
	if middle := strings.TrimSpace(patronymic); middle != "" {
		parts = append(parts, middle)
	}
	if len(parts) == 0 {
		return fallback
	}
	return strings.Join(parts, " ")
}
