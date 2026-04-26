package stage3_ordering

import (
	"fmt"
	"math"
	"math/rand"
)

const RandomSeed = 42

const IterationsPerNode = 1000

func optimizePositions(cm *CoordMatrix) {

	allNodes := []*CoordNode{}
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if !node.IsPseudo {
				allNodes = append(allNodes, node)
			}
		}
	}

	if len(allNodes) == 0 {
		return
	}

	rng := rand.New(rand.NewSource(RandomSeed))

	iterations := IterationsPerNode * len(allNodes)

	initialTemperature := 100.0
	finalTemperature := 0.1

	for iter := 0; iter < iterations; iter++ {

		temperature := initialTemperature - (initialTemperature-finalTemperature)*float64(iter)/float64(iterations)

		node := allNodes[rng.Intn(len(allNodes))]

		if node.IsPseudo {
			continue
		}

		leftNeighbor, rightNeighbor := getNeighbors(cm, node)

		currentDist := calculateTotalDistance(node, cm)

		leftDist := -1
		var leftPushedChain []*CoordNode
		canMoveLeft, pushedLeft := canMoveWithPush(node, -1, leftNeighbor, rightNeighbor, cm)
		if canMoveLeft {

			neighborDistsBefore := make([]int, len(pushedLeft))
			for i, neighbor := range pushedLeft {
				neighborDistsBefore[i] = calculateTotalDistance(neighbor, cm)
			}

			moveNode(node, -1, cm)
			for _, neighbor := range pushedLeft {
				moveNode(neighbor, -1, cm)
			}
			leftPushedChain = pushedLeft

			leftDist = calculateTotalDistance(node, cm)
			for i, neighbor := range pushedLeft {
				neighborDistAfter := calculateTotalDistance(neighbor, cm)
				leftDist += neighborDistAfter - neighborDistsBefore[i]
			}

			moveNode(node, 1, cm)
			for _, neighbor := range pushedLeft {
				moveNode(neighbor, 1, cm)
			}
		}

		rightDist := -1
		var rightPushedChain []*CoordNode
		canMoveRight, pushedRight := canMoveWithPush(node, 1, leftNeighbor, rightNeighbor, cm)
		if canMoveRight {

			neighborDistsBefore := make([]int, len(pushedRight))
			for i, neighbor := range pushedRight {
				neighborDistsBefore[i] = calculateTotalDistance(neighbor, cm)
			}

			moveNode(node, 1, cm)
			for _, neighbor := range pushedRight {
				moveNode(neighbor, 1, cm)
			}
			rightPushedChain = pushedRight

			rightDist = calculateTotalDistance(node, cm)
			for i, neighbor := range pushedRight {
				neighborDistAfter := calculateTotalDistance(neighbor, cm)
				rightDist += neighborDistAfter - neighborDistsBefore[i]
			}

			moveNode(node, -1, cm)
			for _, neighbor := range pushedRight {
				moveNode(neighbor, -1, cm)
			}
		}

		bestDist := currentDist
		bestMove := 0
		var bestPushedChain []*CoordNode = nil

		if canMoveLeft && leftDist < bestDist {
			bestDist = leftDist
			bestMove = -1
			bestPushedChain = leftPushedChain
		}
		if canMoveRight && rightDist < bestDist {
			bestDist = rightDist
			bestMove = 1
			bestPushedChain = rightPushedChain
		}

		if bestMove == 0 && (canMoveLeft || canMoveRight) {

			var possibleMoves []int
			var possibleDists []int
			var possibleChains [][]*CoordNode

			if canMoveLeft {
				possibleMoves = append(possibleMoves, -1)
				possibleDists = append(possibleDists, leftDist)
				possibleChains = append(possibleChains, leftPushedChain)
			}
			if canMoveRight {
				possibleMoves = append(possibleMoves, 1)
				possibleDists = append(possibleDists, rightDist)
				possibleChains = append(possibleChains, rightPushedChain)
			}

			idx := rng.Intn(len(possibleMoves))
			candidateMove := possibleMoves[idx]
			candidateDist := possibleDists[idx]
			candidateChain := possibleChains[idx]

			delta := float64(candidateDist - currentDist)
			if delta <= 0 {

				bestMove = candidateMove
				bestDist = candidateDist
				bestPushedChain = candidateChain
			} else {

				probability := math.Exp(-delta / temperature)
				if rng.Float64() < probability {
					bestMove = candidateMove
					bestDist = candidateDist
					bestPushedChain = candidateChain
				}
			}
		}

		if iter >= iterations-10 {
			nodeName := "?"
			if len(node.People) > 0 {
				nodeName = fmt.Sprintf("%s(ID:%d)", node.People[0].Name, node.People[0].ID)
			}
			fmt.Printf("РС‚РµСЂР°С†РёСЏ %d/%d (T=%.2f): Р’РµСЂС€РёРЅР° [%s] Layer=%d, Left=%d, Right=%d\n",
				iter+1, iterations, temperature, nodeName, node.Layer, node.Left, node.Right)
			fmt.Printf("  РўРµРєСѓС‰РµРµ СЂР°СЃСЃС‚РѕСЏРЅРёРµ: %d\n", currentDist)
			if canMoveLeft {
				fmt.Printf("  РЎРґРІРёРі РІР»РµРІРѕ (-1):   %d (С‚РѕР»РєР°РµРј %d РІРµСЂС€РёРЅ)\n", leftDist, len(leftPushedChain))
			} else {
				fmt.Printf("  РЎРґРІРёРі РІР»РµРІРѕ (-1):   РЅРµРІРѕР·РјРѕР¶РµРЅ\n")
			}
			if canMoveRight {
				fmt.Printf("  РЎРґРІРёРі РІРїСЂР°РІРѕ (+1):  %d (С‚РѕР»РєР°РµРј %d РІРµСЂС€РёРЅ)\n", rightDist, len(rightPushedChain))
			} else {
				fmt.Printf("  РЎРґРІРёРі РІРїСЂР°РІРѕ (+1):  РЅРµРІРѕР·РјРѕР¶РµРЅ\n")
			}
			if bestMove != 0 {
				fmt.Printf("  в†’ РџСЂРёРјРµРЅСЏРµРј СЃРґРІРёРі %+d (РЅРѕРІРѕРµ СЂР°СЃСЃС‚РѕСЏРЅРёРµ: %d)\n", bestMove, bestDist)
			} else {
				fmt.Printf("  в†’ РћСЃС‚Р°С‘РјСЃСЏ РЅР° РјРµСЃС‚Рµ\n")
			}
			fmt.Println()
		}

		if bestMove != 0 {
			moveNode(node, bestMove, cm)
			for _, neighbor := range bestPushedChain {
				moveNode(neighbor, bestMove, cm)
			}
		}
	}
}

func getNeighbors(cm *CoordMatrix, node *CoordNode) (*CoordNode, *CoordNode) {
	var leftNeighbor, rightNeighbor *CoordNode

	nodes := cm.Layers[node.Layer]
	for _, n := range nodes {
		if n == node {
			continue
		}

		if n.Right <= node.Left {
			if leftNeighbor == nil || n.Right > leftNeighbor.Right {
				leftNeighbor = n
			}
		}

		if n.Left >= node.Right {
			if rightNeighbor == nil || n.Left < rightNeighbor.Left {
				rightNeighbor = n
			}
		}
	}

	return leftNeighbor, rightNeighbor
}

func canMove(node *CoordNode, delta int, leftNeighbor, rightNeighbor *CoordNode, cm *CoordMatrix) bool {
	newLeft := node.Left + delta
	newRight := node.Right + delta

	if leftNeighbor != nil && newLeft < leftNeighbor.Right {
		return false
	}
	if rightNeighbor != nil && newRight > rightNeighbor.Left {
		return false
	}

	for _, up := range node.Up {
		if up != nil && up.IsPseudo {
			if !canMovePseudoChain(up, delta, cm) {
				return false
			}
		}
	}

	return true
}

func canMoveWithPush(node *CoordNode, delta int, leftNeighbor, rightNeighbor *CoordNode, cm *CoordMatrix) (bool, []*CoordNode) {

	var pushedChain []*CoordNode
	visited := make(map[*CoordNode]bool)
	visited[node] = true

	canPush := canPushChain(node, delta, leftNeighbor, rightNeighbor, cm, &pushedChain, visited)
	if !canPush {
		return false, nil
	}

	for _, up := range node.Up {
		if up != nil && up.IsPseudo {
			if !canMovePseudoChain(up, delta, cm) {
				return false, nil
			}
		}
	}

	return true, pushedChain
}

func canPushChain(node *CoordNode, delta int, leftNeighbor, rightNeighbor *CoordNode, cm *CoordMatrix, pushedChain *[]*CoordNode, visited map[*CoordNode]bool) bool {
	newLeft := node.Left + delta
	newRight := node.Right + delta

	var blockingNeighbor *CoordNode = nil

	if delta < 0 && leftNeighbor != nil && newLeft < leftNeighbor.Right {

		if leftNeighbor.IsPseudo {
			return false
		}
		blockingNeighbor = leftNeighbor
	}
	if delta > 0 && rightNeighbor != nil && newRight > rightNeighbor.Left {

		if rightNeighbor.IsPseudo {
			return false
		}
		blockingNeighbor = rightNeighbor
	}

	if blockingNeighbor == nil {
		return true
	}

	if visited[blockingNeighbor] {
		return false
	}
	visited[blockingNeighbor] = true

	*pushedChain = append(*pushedChain, blockingNeighbor)

	for _, up := range blockingNeighbor.Up {
		if up != nil && up.IsPseudo {
			if !canMovePseudoChain(up, delta, cm) {
				return false
			}
		}
	}

	neighborLeftN, neighborRightN := getNeighbors(cm, blockingNeighbor)

	if visited[neighborLeftN] {
		neighborLeftN = nil
	}
	if visited[neighborRightN] {
		neighborRightN = nil
	}

	return canPushChain(blockingNeighbor, delta, neighborLeftN, neighborRightN, cm, pushedChain, visited)
}

func canMovePseudoChain(pseudo *CoordNode, delta int, cm *CoordMatrix) bool {
	current := pseudo
	for current != nil && current.IsPseudo {
		leftN, rightN := getNeighbors(cm, current)
		newLeft := current.Left + delta
		newRight := current.Right + delta

		if leftN != nil && newLeft < leftN.Right {
			return false
		}
		if rightN != nil && newRight > rightN.Left {
			return false
		}

		if len(current.Up) > 0 {
			current = current.Up[0]
		} else {
			break
		}
	}
	return true
}

func moveNode(node *CoordNode, delta int, cm *CoordMatrix) {
	node.Left += delta
	node.Right += delta

	for _, up := range node.Up {
		if up != nil && up.IsPseudo {
			movePseudoChainByDelta(up, delta)
		}
	}
}

func movePseudoChainByDelta(pseudo *CoordNode, delta int) {
	current := pseudo
	for current != nil && current.IsPseudo {
		current.Left += delta
		current.Right += delta
		if len(current.Up) > 0 {
			current = current.Up[0]
		} else {
			break
		}
	}
}
