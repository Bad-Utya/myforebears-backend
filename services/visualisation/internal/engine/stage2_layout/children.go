package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func AddChildren(mainParent, otherParent *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	children := GetCommonChildren(mainParent, otherParent)

	if len(children) == 0 {
		return
	}

	var childDirection stage1_input.PlacementDirection
	var leftParent, rightParent *stage1_input.Person

	otherIsLeft := isPartnerOnLeft(mainParent, otherParent)

	if otherIsLeft {

		childDirection = stage1_input.PlacedLeft
		leftParent = otherParent
		rightParent = mainParent
	} else {

		childDirection = stage1_input.PlacedRight
		leftParent = mainParent
		rightParent = otherParent
	}

	childLayer := mainParent.Layout.Layer - 1

	var iterOrder []int
	if otherIsLeft {

		for i := len(children) - 1; i >= 0; i-- {
			iterOrder = append(iterOrder, i)
		}
	} else {

		for i := 0; i < len(children); i++ {
			iterOrder = append(iterOrder, i)
		}
	}

	for _, i := range iterOrder {
		child := children[i]

		if child.IsLayouted() {
			continue
		}

		child.Layout = stage1_input.NewPersonLayout(childLayer)

		child.Layout.AddedFromLeft = (childDirection == stage1_input.PlacedLeft)

		isFirst := (i == 0)
		isLast := (i == len(children)-1)

		if isFirst && isLast {

			child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
			child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
		} else if isFirst {

			child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
			child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
		} else if isLast {

			child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
			child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
		} else {

			child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
			child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
		}

		queue.Enqueue(child)

		history.Add(otherParent, child, -1, childDirection, stage1_input.RelationChild)
	}
}

func isPartnerOnLeft(mainParent, otherParent *stage1_input.Person) bool {

	if mainParent.Layout.DirectionConstraint == stage1_input.OnlyLeft {

		return true
	}
	if mainParent.Layout.DirectionConstraint == stage1_input.OnlyRight {

		return false
	}

	return false
}
