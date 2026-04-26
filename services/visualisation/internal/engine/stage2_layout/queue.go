package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// Queue РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РѕС‡РµСЂРµРґСЊ РґР»СЏ BFS РѕР±С…РѕРґР°
type Queue struct {
	items []*stage1_input.Person
}

// NewQueue СЃРѕР·РґР°С‘С‚ РЅРѕРІСѓСЋ РїСѓСЃС‚СѓСЋ РѕС‡РµСЂРµРґСЊ
func NewQueue() *Queue {
	return &Queue{
		items: make([]*stage1_input.Person, 0),
	}
}

// Enqueue РґРѕР±Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° РІ РєРѕРЅРµС† РѕС‡РµСЂРµРґРё
func (q *Queue) Enqueue(person *stage1_input.Person) {
	q.items = append(q.items, person)
}

// Dequeue РёР·РІР»РµРєР°РµС‚ С‡РµР»РѕРІРµРєР° РёР· РЅР°С‡Р°Р»Р° РѕС‡РµСЂРµРґРё
func (q *Queue) Dequeue() *stage1_input.Person {
	if len(q.items) == 0 {
		return nil
	}
	person := q.items[0]
	q.items = q.items[1:]
	return person
}

// IsEmpty РїСЂРѕРІРµСЂСЏРµС‚, РїСѓСЃС‚Р° Р»Рё РѕС‡РµСЂРµРґСЊ
func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

// Size РІРѕР·РІСЂР°С‰Р°РµС‚ СЂР°Р·РјРµСЂ РѕС‡РµСЂРµРґРё
func (q *Queue) Size() int {
	return len(q.items)
}

// QueueEntry rappresenta un elemento nella coda con tracciamento della profondità
type QueueEntry struct {
	Person *stage1_input.Person
	Depth  int
}

// QueueWithDepth представляет очередь для BFS обхода с отслеживанием глубины
type QueueWithDepth struct {
	items []*QueueEntry
}

// NewQueueWithDepth создаёт новую пустую очередь с отслеживанием глубины
func NewQueueWithDepth() *QueueWithDepth {
	return &QueueWithDepth{
		items: make([]*QueueEntry, 0),
	}
}

// Enqueue добавляет человека с глубиной в конец очереди
func (q *QueueWithDepth) Enqueue(person *stage1_input.Person, depth int) {
	q.items = append(q.items, &QueueEntry{Person: person, Depth: depth})
}

// DequeueWithDepth извлекает человека и его глубину из начала очереди
func (q *QueueWithDepth) DequeueWithDepth() (*stage1_input.Person, int) {
	if len(q.items) == 0 {
		return nil, 0
	}
	entry := q.items[0]
	q.items = q.items[1:]
	return entry.Person, entry.Depth
}

// IsEmpty проверяет, пуста ли очередь
func (q *QueueWithDepth) IsEmpty() bool {
	return len(q.items) == 0
}

// Size возвращает размер очереди
func (q *QueueWithDepth) Size() int {
	return len(q.items)
}
