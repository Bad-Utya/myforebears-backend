package stage1_input

type DirectionConstraint int

const (
	NoDirectionConstraint DirectionConstraint = iota
	OnlyLeft
	OnlyRight
)

type HeightConstraint struct {
	MaxLayer *int
	CausedBy *Person
}

type PersonLayout struct {
	Layer int

	LeftHeightConstraint  *HeightConstraint
	RightHeightConstraint *HeightConstraint

	DirectionConstraint DirectionConstraint

	AddedFromLeft bool

	IsStartPerson bool

	Processed bool
}

func NewPersonLayout(layer int) *PersonLayout {
	return &PersonLayout{
		Layer:     layer,
		Processed: false,
	}
}

func CopyHeightConstraint(constraint *HeightConstraint) *HeightConstraint {
	if constraint == nil {
		return nil
	}
	return &HeightConstraint{
		MaxLayer: constraint.MaxLayer,
		CausedBy: constraint.CausedBy,
	}
}

func NewHeightConstraint(layerPtr *int, causedBy *Person) *HeightConstraint {
	return &HeightConstraint{
		MaxLayer: layerPtr,
		CausedBy: causedBy,
	}
}

func (c *HeightConstraint) CanAddAbove(currentLayer int) (bool, *Person) {
	if c == nil || c.MaxLayer == nil {
		return true, nil
	}

	parentLayer := currentLayer + 1
	if parentLayer >= *c.MaxLayer {
		return false, c.CausedBy
	}
	return true, nil
}
