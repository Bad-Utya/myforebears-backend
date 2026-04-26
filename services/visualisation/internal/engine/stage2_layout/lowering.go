package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func LowerSubtree(person *stage1_input.Person, isInitial bool, visited VisitedSet) {

	if visited.Contains(person.ID) {
		return
	}

	if person.Layout == nil {
		return
	}

	visited.Add(person.ID)

	person.Layout.Layer--

	for _, partner := range person.Partners {
		if partner.Layout != nil {
			LowerSubtree(partner, false, visited)
		}
	}

	for _, child := range person.Children {
		if child.Layout != nil {
			LowerSubtree(child, false, visited)
		}
	}

	if !isInitial {
		if person.Mother != nil && person.Mother.Layout != nil {
			LowerSubtree(person.Mother, false, visited)
		}
		if person.Father != nil && person.Father.Layout != nil {
			LowerSubtree(person.Father, false, visited)
		}
	}
}
