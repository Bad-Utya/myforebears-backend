package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func AddPartners(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	for _, partner := range person.Partners {

		addLeft := ShouldAddPartnerLeft(person, partner, person.Layout.DirectionConstraint)

		if !partner.IsLayouted() {
			partner.Layout = stage1_input.NewPersonLayout(person.Layout.Layer)

			partner.Layout.AddedFromLeft = addLeft

			if addLeft {

				partner.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
				partner.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
			} else {

				partner.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
				partner.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
			}

			if person.HasParents() {
				if addLeft {
					partner.Layout.DirectionConstraint = stage1_input.OnlyLeft
				} else {
					partner.Layout.DirectionConstraint = stage1_input.OnlyRight
				}
			}

			queue.Enqueue(partner)

			direction := stage1_input.PlacedRight
			if addLeft {
				direction = stage1_input.PlacedLeft
			}
			history.Add(person, partner, 0, direction, stage1_input.RelationPartner)
		}

		AddChildren(person, partner, queue, history)
	}
}
