package stage1_input

type Gender int

const (
	Male Gender = iota
	Female
)

type Person struct {
	ID     int
	Name   string
	Gender Gender

	Mother   *Person
	Father   *Person
	Partners []*Person
	Children []*Person

	Layout *PersonLayout
}

func NewPerson(id int, name string, gender Gender) *Person {
	return &Person{
		ID:       id,
		Name:     name,
		Gender:   gender,
		Partners: make([]*Person, 0),
		Children: make([]*Person, 0),
	}
}

func (p *Person) HasParents() bool {
	return p.Mother != nil && p.Father != nil
}

func (p *Person) IsLayouted() bool {
	return p.Layout != nil
}

func (p *Person) IsProcessed() bool {
	return p.Layout != nil && p.Layout.Processed
}
