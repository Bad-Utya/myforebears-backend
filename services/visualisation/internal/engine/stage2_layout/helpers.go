package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

type VisitedSet map[int]bool

func NewVisitedSet() VisitedSet {
	return make(VisitedSet)
}

func (v VisitedSet) Add(id int) {
	v[id] = true
}

func (v VisitedSet) Contains(id int) bool {
	return v[id]
}

func GetCommonChildren(parent1, parent2 *stage1_input.Person) []*stage1_input.Person {
	result := make([]*stage1_input.Person, 0)

	parent2ChildrenIDs := make(map[int]bool)
	for _, child := range parent2.Children {
		parent2ChildrenIDs[child.ID] = true
	}

	for _, child := range parent1.Children {
		if parent2ChildrenIDs[child.ID] {
			result = append(result, child)
		}
	}

	return result
}

func IsCurrentPartner(person, partner *stage1_input.Person) bool {
	if len(person.Partners) == 0 {
		return false
	}

	bestPartner := person.Partners[0]
	bestScore := countSharedChildren(person, bestPartner)

	for _, candidate := range person.Partners[1:] {
		score := countSharedChildren(person, candidate)
		if score > bestScore || (score == bestScore && candidate.ID < bestPartner.ID) {
			bestPartner = candidate
			bestScore = score
		}
	}

	return bestPartner.ID == partner.ID
}

func ShouldAddPartnerLeft(person, partner *stage1_input.Person, dirConstraint stage1_input.DirectionConstraint) bool {

	if dirConstraint == stage1_input.OnlyLeft {
		return true
	}
	if dirConstraint == stage1_input.OnlyRight {
		return false
	}

	isCurrent := IsCurrentPartner(person, partner)

	if person.Gender == stage1_input.Male {

		if isCurrent {
			return false
		}
		return true
	} else {

		if isCurrent {
			return true
		}
		return false
	}
}

func countSharedChildren(person, partner *stage1_input.Person) int {
	if person == nil || partner == nil {
		return 0
	}

	return len(GetCommonChildren(person, partner))
}
