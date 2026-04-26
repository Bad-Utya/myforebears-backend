package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

type CoordNode struct {
	Left  int
	Right int

	People []*stage1_input.Person

	IsPseudo bool

	Up []*CoordNode

	Down []*CoordNode

	Layer int

	OriginalNode *LayerNode

	WasMerged bool

	MergePartner *CoordNode

	ParentNodes []*CoordNode

	ParentPersonIndex []int

	Children []*CoordNode

	AddedLeft bool
}

func (cn *CoordNode) Width() int {
	return cn.Right - cn.Left
}

func (cn *CoordNode) Center() int {
	return (cn.Left + cn.Right) / 2
}
