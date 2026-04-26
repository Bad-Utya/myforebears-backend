package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func ProcessPlacementHistory(
	history *stage1_input.PlacementHistory,
	startPerson *stage1_input.Person,
	startLayer int,
	layouts map[int]*stage1_input.PersonLayout,
) *OrderManager {

	minLayer, maxLayer := startLayer, startLayer
	for _, layout := range layouts {
		if layout.Layer < minLayer {
			minLayer = layout.Layer
		}
		if layout.Layer > maxLayer {
			maxLayer = layout.Layer
		}
	}

	om := NewOrderManager(minLayer, maxLayer)

	om.AddStartPerson(startPerson, startLayer)

	for _, record := range history.Records {
		fromPerson := record.FromPerson
		addedPerson := record.AddedPerson
		addedPerson2 := record.AddedPerson2
		direction := record.Direction
		relationType := record.RelationType

		fromLayout := layouts[fromPerson.ID]
		addedLayout := layouts[addedPerson.ID]
		if fromLayout == nil || addedLayout == nil {
			continue
		}

		fromLayer := fromLayout.Layer
		targetLayer := addedLayout.Layer

		fromNode := om.GetPersonNode(fromPerson.ID)
		if fromNode == nil {
			continue
		}

		if addedPerson2 != nil && relationType == stage1_input.RelationParent {

			var siblingUp *LayerNode
			var fromPersonIndex int = -1

			for i, p := range fromNode.People {
				if p.ID == fromPerson.ID {
					fromPersonIndex = i
					break
				}
			}

			if len(fromNode.People) == 2 && fromPersonIndex >= 0 {
				otherIndex := 1 - fromPersonIndex
				if otherIndex < len(fromNode.Up) && fromNode.Up[otherIndex] != nil {
					siblingUp = fromNode.Up[otherIndex]
				}
			}

			if direction == stage1_input.PlacedLeft {
				om.AddParentPairLeft(fromNode, addedPerson, addedPerson2, fromLayer, targetLayer, siblingUp, fromPersonIndex)
			} else {
				om.AddParentPairRight(fromNode, addedPerson, addedPerson2, fromLayer, targetLayer, siblingUp, fromPersonIndex)
			}
			continue
		}

		var fromPersonIndex int = -1
		for i, p := range fromNode.People {
			if p.ID == fromPerson.ID {
				fromPersonIndex = i
				break
			}
		}
		if fromPersonIndex < 0 {
			fromPersonIndex = 0
		}

		if direction == stage1_input.PlacedLeft {
			om.AddPersonLeft(fromNode, addedPerson, fromLayer, targetLayer, fromPersonIndex)
		} else {
			om.AddPersonRight(fromNode, addedPerson, fromLayer, targetLayer, fromPersonIndex)
		}
	}

	return om
}
