package engine

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage2_layout"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage4_render"
)

func RenderCoordinates(
	visType models.VisualisationType,
	rootPersonID uuid.UUID,
	includedPersonIDs []uuid.UUID,
	content *familytreepb.GetTreeContentResponse,
	maxDepth int,
	allowLayerShift bool,
) ([]byte, error) {
	return RenderCoordinatesWithTrace(visType, rootPersonID, includedPersonIDs, content, maxDepth, allowLayerShift, nil)
}

func RenderCoordinatesWithTrace(
	visType models.VisualisationType,
	rootPersonID uuid.UUID,
	includedPersonIDs []uuid.UUID,
	content *familytreepb.GetTreeContentResponse,
	maxDepth int,
	allowLayerShift bool,
	out io.Writer,
) ([]byte, error) {
	return renderCoordinatesInternal(visType, rootPersonID, includedPersonIDs, content, maxDepth, allowLayerShift, out)
}

func renderCoordinatesInternal(
	visType models.VisualisationType,
	rootPersonID uuid.UUID,
	includedPersonIDs []uuid.UUID,
	content *familytreepb.GetTreeContentResponse,
	maxDepth int,
	allowLayerShift bool,
	out io.Writer,
) ([]byte, error) {
	if content == nil {
		return nil, fmt.Errorf("visualisation content is empty")
	}
	debugf(out, "=== CLIENT RENDER (COORDINATES) ===")
	debugf(out, "input: vis_type=%s root=%s included=%d max_depth=%d allow_layer_shift=%v",
		visType, rootPersonID.String(), len(includedPersonIDs), maxDepth, allowLayerShift)

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
	if visType == models.VisualisationTypeFull && len(includedPersonIDs) == 0 {
		allowed = make(map[uuid.UUID]struct{}, len(people))
		for id := range people {
			allowed[id] = struct{}{}
		}
	}

	filteredPeople, filteredRelations := filterByVisualisation(visType, rootPersonID, allowed, people, relations)
	debugf(out, "[filter] people=%d relations=%d", len(filteredPeople), len(filteredRelations))
	if out != nil {
		debugPrintPeople(out, filteredPeople)
		debugPrintRelations(out, filteredRelations)
	}
	if _, ok := filteredPeople[rootPersonID]; !ok {
		return nil, errRootNotFound
	}

	tree, idToInt, internalPersons, err := buildStageTree(filteredPeople, filteredRelations)
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

	var history *stage1_input.PlacementHistory
	if maxDepth > 0 {
		history, err = stage2_layout.LayoutFromPersonWithDepthSimple(tree, rootIntID, maxDepth)
		debugf(out, "[stage2_layout] with maxDepth=%d", maxDepth)
	} else {
		history, err = stage2_layout.LayoutFromPerson(tree, rootIntID)
		debugf(out, "[stage2_layout] unlimited depth")
	}
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
	if !allowLayerShift {
		debugf(out, "[stage3_ordering] layer shift disabled (flag set but not yet implemented)")

	}
	if out != nil {
		debugPrintOrder(out, om)
	}
	cm := om.BuildCoordMatrix()
	if out != nil {
		debugPrintCoordMatrix(out, cm)
	}
	renderResult := stage4_render.BuildCoordRenderResult(cm, tree)
	debugf(out, "[stage4_render] nodes=%d edges=%d", len(renderResult.Nodes), len(renderResult.Edges))

	return renderResultToJSON(renderResult, internalPersons), nil
}

type CoordinateNodeJSON struct {
	ID         int          `json:"id"`
	X          int          `json:"x"`
	Y          int          `json:"y"`
	People     []PersonJSON `json:"people"`
	PartnerIdx int          `json:"partnerIdx"`
	MergeLeft  bool         `json:"mergeLeft"`
}

type PersonJSON struct {
	ID            string `json:"id"`
	TreeID        string `json:"tree_id,omitempty"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	Patronymic    string `json:"patronymic,omitempty"`
	Gender        string `json:"gender"`
	AvatarPhotoID string `json:"avatar_photo_id,omitempty"`
	Name          string `json:"name,omitempty"`
}

type CoordinateEdgeJSON struct {
	FromNodeIdx int    `json:"fromNodeIdx"`
	ToNodeIdx   int    `json:"toNodeIdx"`
	FromX       int    `json:"fromX"`
	FromY       int    `json:"fromY"`
	ToX         int    `json:"toX"`
	ToY         int    `json:"toY"`
	EdgeType    string `json:"edgeType"`
}

type CoordinateResultJSON struct {
	Nodes    []CoordinateNodeJSON `json:"nodes"`
	Edges    []CoordinateEdgeJSON `json:"edges"`
	MinLayer int                  `json:"minLayer"`
	MaxLayer int                  `json:"maxLayer"`
	MaxRight int                  `json:"maxRight"`
}

func renderResultToJSON(result *stage4_render.CoordRenderResult, internalPersons map[int]familytreepb.Person) []byte {
	if result == nil {
		data, _ := json.Marshal(CoordinateResultJSON{
			Nodes:    []CoordinateNodeJSON{},
			Edges:    []CoordinateEdgeJSON{},
			MinLayer: 0,
			MaxLayer: 0,
			MaxRight: 0,
		})
		return data
	}

	nodes := make([]CoordinateNodeJSON, len(result.Nodes))
	for i, nodeInfo := range result.Nodes {
		peopleJSON := make([]PersonJSON, len(nodeInfo.People))
		for j, person := range nodeInfo.People {
			fullPerson, ok := internalPersons[person.ID]
			gender := "unknown"
			if person.Gender == stage1_input.Male {
				gender = "male"
			} else if person.Gender == stage1_input.Female {
				gender = "female"
			}
			if ok {
				peopleJSON[j] = PersonJSON{
					ID:            fullPerson.GetId(),
					TreeID:        fullPerson.GetTreeId(),
					FirstName:     fullPerson.GetFirstName(),
					LastName:      fullPerson.GetLastName(),
					Patronymic:    fullPerson.GetPatronymic(),
					Gender:        gender,
					AvatarPhotoID: fullPerson.GetAvatarPhotoId(),
					Name:          buildPersonDisplayName(fullPerson.GetFirstName(), fullPerson.GetLastName(), fullPerson.GetPatronymic(), fullPerson.GetId()),
				}
				continue
			}
			peopleJSON[j] = PersonJSON{
				ID:     fmt.Sprintf("%d", person.ID),
				Name:   person.Name,
				Gender: gender,
			}
		}

		nodes[i] = CoordinateNodeJSON{
			ID:         i,
			X:          nodeInfo.Left,
			Y:          nodeInfo.Layer,
			People:     peopleJSON,
			PartnerIdx: nodeInfo.MergePartnerIdx,
			MergeLeft:  nodeInfo.AddedLeft,
		}
	}

	edges := make([]CoordinateEdgeJSON, len(result.Edges))
	for i, edgeInfo := range result.Edges {
		edges[i] = CoordinateEdgeJSON{
			FromNodeIdx: edgeInfo.FromNodeIdx,
			ToNodeIdx:   edgeInfo.ToNodeIdx,
			FromX:       edgeInfo.FromX,
			FromY:       edgeInfo.FromY,
			ToX:         edgeInfo.ToX,
			ToY:         edgeInfo.ToY,
			EdgeType:    edgeInfo.EdgeType,
		}
	}

	coordResult := CoordinateResultJSON{
		Nodes:    nodes,
		Edges:    edges,
		MinLayer: result.MinLayer,
		MaxLayer: result.MaxLayer,
		MaxRight: result.MaxRight,
	}

	data, _ := json.MarshalIndent(coordResult, "", "  ")
	return data
}
