package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

type PersonPosition struct {
	Person *stage1_input.Person
	X      int
	Y      int
}

type PeopleGrid struct {
	Layers map[int][]PersonPosition

	Positions []PersonPosition
}

func (om *OrderManager) BuildPeopleGrid() *PeopleGrid {
	grid := &PeopleGrid{
		Layers:    make(map[int][]PersonPosition),
		Positions: []PersonPosition{},
	}

	for _, layer := range om.GetAllLayers() {
		positions := []PersonPosition{}
		nodes := layer.GetNodes()
		xIndex := 0

		for _, node := range nodes {

			if node.IsPseudo {
				continue
			}

			for _, person := range node.People {
				pos := PersonPosition{
					Person: person,
					X:      xIndex,
					Y:      layer.Number,
				}
				positions = append(positions, pos)
				grid.Positions = append(grid.Positions, pos)
				xIndex++
			}
		}

		grid.Layers[layer.Number] = positions
	}

	return grid
}
