package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

type Queue struct {
	items []*stage1_input.Person
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]*stage1_input.Person, 0),
	}
}

func (q *Queue) Enqueue(person *stage1_input.Person) {
	q.items = append(q.items, person)
}

func (q *Queue) Dequeue() *stage1_input.Person {
	if len(q.items) == 0 {
		return nil
	}
	person := q.items[0]
	q.items = q.items[1:]
	return person
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *Queue) Size() int {
	return len(q.items)
}

type QueueEntry struct {
	Person *stage1_input.Person
	Depth  int
}

type QueueWithDepth struct {
	items []*QueueEntry
}

func NewQueueWithDepth() *QueueWithDepth {
	return &QueueWithDepth{
		items: make([]*QueueEntry, 0),
	}
}

func (q *QueueWithDepth) Enqueue(person *stage1_input.Person, depth int) {
	q.items = append(q.items, &QueueEntry{Person: person, Depth: depth})
}

func (q *QueueWithDepth) DequeueWithDepth() (*stage1_input.Person, int) {
	if len(q.items) == 0 {
		return nil, 0
	}
	entry := q.items[0]
	q.items = q.items[1:]
	return entry.Person, entry.Depth
}

func (q *QueueWithDepth) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *QueueWithDepth) Size() int {
	return len(q.items)
}
