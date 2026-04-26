package stage2_layout

import (
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

func LayoutFromPerson(tree *stage1_input.FamilyTree, startPersonID int) (*stage1_input.PlacementHistory, error) {
	startPerson := tree.GetPerson(startPersonID)
	if startPerson == nil {
		return nil, fmt.Errorf("person with ID %d not found", startPersonID)
	}

	history := stage1_input.NewPlacementHistory()

	startPerson.Layout = stage1_input.NewPersonLayout(0)
	startPerson.Layout.IsStartPerson = true

	queue := NewQueue()
	queue.Enqueue(startPerson)

	for !queue.IsEmpty() {
		person := queue.Dequeue()

		if person.IsProcessed() {
			continue
		}

		processPersonLayout(person, queue, history)

		person.Layout.Processed = true
	}

	return history, nil
}

func processPersonLayout(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {

	for !AddParents(person, queue, history) {

	}

	AddPartners(person, queue, history)
}
