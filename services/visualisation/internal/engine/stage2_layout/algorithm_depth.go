package stage2_layout

import (
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

func LayoutFromPersonWithDepth(tree *stage1_input.FamilyTree, startPersonID int, maxDepth int) (*stage1_input.PlacementHistory, error) {
	if maxDepth == 0 {

		return LayoutFromPerson(tree, startPersonID)
	}

	startPerson := tree.GetPerson(startPersonID)
	if startPerson == nil {
		return nil, fmt.Errorf("person with ID %d not found", startPersonID)
	}

	history := stage1_input.NewPlacementHistory()

	startPerson.Layout = stage1_input.NewPersonLayout(0)
	startPerson.Layout.IsStartPerson = true

	queueDepths := make(map[*stage1_input.Person]int)
	queue := NewQueue()
	queue.Enqueue(startPerson)
	queueDepths[startPerson] = 0

	for !queue.IsEmpty() {
		person := queue.Dequeue()
		depth := queueDepths[person]

		if person.IsProcessed() {
			continue
		}

		if depth >= maxDepth {

			person.Layout.Processed = true
			delete(queueDepths, person)
			continue
		}

		processPersonLayout(person, queue, history)

		person.Layout.Processed = true
		delete(queueDepths, person)

		updateDepthsInQueue(tree, queue, queueDepths, depth, maxDepth)
	}

	return history, nil
}

func updateDepthsInQueue(
	tree *stage1_input.FamilyTree,
	queue *Queue,
	depths map[*stage1_input.Person]int,
	currentDepth int,
	maxDepth int,
) {

}

func LayoutFromPersonWithDepthSimple(tree *stage1_input.FamilyTree, startPersonID int, maxDepth int) (*stage1_input.PlacementHistory, error) {
	if maxDepth == 0 {
		return LayoutFromPerson(tree, startPersonID)
	}

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

		processPersonLayoutWithDepthFilter(person, queue, history, maxDepth)

		person.Layout.Processed = true
	}

	return history, nil
}

func processPersonLayoutWithDepthFilter(
	person *stage1_input.Person,
	queue *Queue,
	history *stage1_input.PlacementHistory,
	maxDepth int,
) {

	personLayer := person.Layout.Layer

	if personLayer+1 < maxDepth {
		for !AddParents(person, queue, history) {

		}
	}

	if personLayer < maxDepth {
		AddPartners(person, queue, history)
	}

}
