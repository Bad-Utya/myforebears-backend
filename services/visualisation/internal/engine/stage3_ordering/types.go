package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

type LayerNode struct {
	People []*stage1_input.Person

	IsPseudo bool

	Prev *LayerNode
	Next *LayerNode

	Up []*LayerNode

	LeftDown  *LayerNode
	RightDown *LayerNode

	Layer int

	AddedLeft bool
}

func (n *LayerNode) IsPerson() bool {
	return len(n.People) > 0 && !n.IsPseudo
}

func (n *LayerNode) HasPerson(personID int) bool {
	for _, p := range n.People {
		if p.ID == personID {
			return true
		}
	}
	return false
}

func (n *LayerNode) GetLeftUp() *LayerNode {
	for i := 0; i < len(n.Up); i++ {
		if n.Up[i] != nil {
			return n.Up[i]
		}
	}
	return nil
}

func (n *LayerNode) GetRightUp() *LayerNode {
	for i := len(n.Up) - 1; i >= 0; i-- {
		if n.Up[i] != nil {
			return n.Up[i]
		}
	}
	return nil
}

func (n *LayerNode) AddPersonToNode(person *stage1_input.Person, position string) {
	if position == "left" {
		n.People = append([]*stage1_input.Person{person}, n.People...)

		n.Up = append([]*LayerNode{nil}, n.Up...)
	} else {
		n.People = append(n.People, person)
	}
}

func (n *LayerNode) IsHead() bool {
	return n.Prev == nil
}

func (n *LayerNode) IsTail() bool {
	return n.Next == nil
}

type Layer struct {
	Number int
	Head   *LayerNode
	Tail   *LayerNode
}

func (l *Layer) GetNodes() []*LayerNode {
	var nodes []*LayerNode
	for node := l.Head.Next; node != nil && node != l.Tail; node = node.Next {
		nodes = append(nodes, node)
	}
	return nodes
}

func (l *Layer) GetPeople() []*stage1_input.Person {
	var people []*stage1_input.Person
	for node := l.Head.Next; node != nil && node != l.Tail; node = node.Next {
		for _, p := range node.People {
			people = append(people, p)
		}
	}
	return people
}

func (l *Layer) GetPeopleIDs() []int {
	var ids []int
	for node := l.Head.Next; node != nil && node != l.Tail; node = node.Next {
		for _, p := range node.People {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func (l *Layer) IsEmpty() bool {
	return l.Head.Next == l.Tail
}
