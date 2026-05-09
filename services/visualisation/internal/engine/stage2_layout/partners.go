package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func AddPartners(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	partners := orderPartnersForPlacement(person)

	for _, partner := range partners {

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

func orderPartnersForPlacement(person *stage1_input.Person) []*stage1_input.Person {
	if len(person.Partners) <= 1 {
		return person.Partners
	}

	ordered := make([]*stage1_input.Person, len(person.Partners))
	copy(ordered, person.Partners)

	for i := 0; i < len(ordered)-1; i++ {
		for j := i + 1; j < len(ordered); j++ {
			leftScore := partnerPlacementScore(person, ordered[i])
			rightScore := partnerPlacementScore(person, ordered[j])
			if rightScore > leftScore || (rightScore == leftScore && ordered[j].ID < ordered[i].ID) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}

	return ordered
}

func partnerPlacementScore(person, partner *stage1_input.Person) int {
	score := countSharedChildren(person, partner) * 10
	if IsCurrentPartner(person, partner) {
		score += 1000
	}
	return score
}
