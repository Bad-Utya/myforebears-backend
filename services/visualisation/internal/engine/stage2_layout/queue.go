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
