package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func AddParents(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) bool {

	if !person.HasParents() {
		return true
	}

	if person.Mother.IsLayouted() && person.Father.IsLayouted() {
		return true
	}

	parentLayer := person.Layout.Layer + 1

	if person.Layout.LeftHeightConstraint != nil {
		canAdd, causedBy := person.Layout.LeftHeightConstraint.CanAddAbove(person.Layout.Layer)
		if !canAdd && causedBy != nil {

			visited := NewVisitedSet()
			LowerSubtree(causedBy, true, visited)

			return false
		}
	}

	if person.Layout.RightHeightConstraint != nil {
		canAdd, causedBy := person.Layout.RightHeightConstraint.CanAddAbove(person.Layout.Layer)
		if !canAdd && causedBy != nil {

			visited := NewVisitedSet()
			LowerSubtree(causedBy, true, visited)
			return false
		}
	}

	var firstParent, secondParent *stage1_input.Person
	var firstDirection, secondDirection stage1_input.PlacementDirection

	switch person.Layout.DirectionConstraint {
	case stage1_input.OnlyRight:

		firstParent = person.Father
		secondParent = person.Mother
		firstDirection = stage1_input.PlacedRight
		secondDirection = stage1_input.PlacedRight
	case stage1_input.OnlyLeft:

		firstParent = person.Mother
		secondParent = person.Father
		firstDirection = stage1_input.PlacedLeft
		secondDirection = stage1_input.PlacedLeft
	default:

		firstParent = person.Father
		secondParent = person.Mother
		firstDirection = stage1_input.PlacedLeft
		secondDirection = stage1_input.PlacedRight
	}

	bothParentsNew := !firstParent.IsLayouted() && !secondParent.IsLayouted()

	if bothParentsNew {

		firstParent.Layout = stage1_input.NewPersonLayout(parentLayer)
		firstParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
		firstParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
		firstParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
		queue.Enqueue(firstParent)

		secondParent.Layout = stage1_input.NewPersonLayout(parentLayer)
		secondParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
		secondParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
		secondParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
		queue.Enqueue(secondParent)

		history.AddPair(person, firstParent, secondParent, 1, firstDirection, stage1_input.RelationParent)
	} else {

		if !firstParent.IsLayouted() {
			firstParent.Layout = stage1_input.NewPersonLayout(parentLayer)
			firstParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
			firstParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
			firstParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
			queue.Enqueue(firstParent)
			history.Add(person, firstParent, 1, firstDirection, stage1_input.RelationParent)
		}

		if !secondParent.IsLayouted() {
			secondParent.Layout = stage1_input.NewPersonLayout(parentLayer)
			secondParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
			secondParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
			secondParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
			queue.Enqueue(secondParent)
			history.Add(person, secondParent, 1, secondDirection, stage1_input.RelationParent)
		}
	}

	addSiblings(person, firstParent, secondParent, queue, history)

	return true
}

func addSiblings(person *stage1_input.Person, leftParentArg *stage1_input.Person, rightParentArg *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {

	siblings := GetCommonChildren(person.Mother, person.Father)

	if len(siblings) <= 1 {

		return
	}

	addSiblingsRight := false
	personIsFirst := false

	switch person.Layout.DirectionConstraint {
	case stage1_input.OnlyRight:
		addSiblingsRight = true
		personIsFirst = true
	case stage1_input.OnlyLeft:
		addSiblingsRight = false
		personIsFirst = false
	default:
		if person.Layout.IsStartPerson {

			if person.Gender == stage1_input.Male {
				addSiblingsRight = false
				personIsFirst = false
			} else {
				addSiblingsRight = true
				personIsFirst = true
			}
		} else if person.Layout.AddedFromLeft {

			addSiblingsRight = false
			personIsFirst = false
		} else if person.Layout.DirectionConstraint != stage1_input.NoDirectionConstraint {

			addSiblingsRight = true
			personIsFirst = true
		} else {

			if person.Gender == stage1_input.Male {
				addSiblingsRight = false
				personIsFirst = false
			} else {
				addSiblingsRight = true
				personIsFirst = true
			}
		}
	}

	var leftParent, rightParent *stage1_input.Person
	if person.Father.Layout != nil && person.Mother.Layout != nil {

		if person.Layout.DirectionConstraint == stage1_input.OnlyRight {
			leftParent = person.Mother
			rightParent = person.Father
		} else {
			leftParent = person.Father
			rightParent = person.Mother
		}
	} else {
		leftParent = person.Father
		rightParent = person.Mother
	}

	var orderedChildren []*stage1_input.Person
	if personIsFirst {
		orderedChildren = append(orderedChildren, person)
		for _, sibling := range siblings {
			if sibling.ID != person.ID {
				orderedChildren = append(orderedChildren, sibling)
			}
		}
	} else {
		for _, sibling := range siblings {
			if sibling.ID != person.ID {
				orderedChildren = append(orderedChildren, sibling)
			}
		}
		orderedChildren = append(orderedChildren, person)
	}

	var iterOrder []int
	if addSiblingsRight {
		for i := 0; i < len(orderedChildren); i++ {
			iterOrder = append(iterOrder, i)
		}
	} else {
		for i := len(orderedChildren) - 1; i >= 0; i-- {
			iterOrder = append(iterOrder, i)
		}
	}

	for _, idx := range iterOrder {
		child := orderedChildren[idx]
		isFirst := (idx == 0)
		isLast := (idx == len(orderedChildren)-1)

		if child.ID == person.ID {

			updateChildHeightConstraints(child, leftParent, rightParent, isFirst, isLast)
		} else {

			if child.IsLayouted() {
				continue
			}

			child.Layout = stage1_input.NewPersonLayout(person.Layout.Layer)
			child.Layout.DirectionConstraint = person.Layout.DirectionConstraint

			child.Layout.AddedFromLeft = !addSiblingsRight

			updateChildHeightConstraints(child, leftParent, rightParent, isFirst, isLast)

			queue.Enqueue(child)

			dir := stage1_input.PlacedLeft
			var fromParent *stage1_input.Person
			if addSiblingsRight {
				dir = stage1_input.PlacedRight
				fromParent = rightParent
			} else {
				fromParent = leftParent
			}
			history.Add(fromParent, child, -1, dir, stage1_input.RelationChild)
		}
	}
}

func updateChildHeightConstraints(child *stage1_input.Person, leftParent, rightParent *stage1_input.Person, isFirst, isLast bool) {
	if isFirst && isLast {

		child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
		child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
	} else if isFirst {

		child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
		child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&leftParent.Layout.Layer, child)
	} else if isLast {

		child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&rightParent.Layout.Layer, child)
		child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
	} else {

		child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&leftParent.Layout.Layer, child)
		child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&rightParent.Layout.Layer, child)
	}
}
