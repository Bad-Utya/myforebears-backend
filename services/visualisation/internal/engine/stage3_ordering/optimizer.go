package stage3_ordering

import (
	"fmt"
	"math"
	"math/rand"
)

// RandomSeed вЂ” С„РёРєСЃРёСЂРѕРІР°РЅРЅС‹Р№ СЃРёРґ РґР»СЏ РіРµРЅРµСЂР°С‚РѕСЂР° СЃР»СѓС‡Р°Р№РЅС‹С… С‡РёСЃРµР»
const RandomSeed = 42

// IterationsPerNode вЂ” РєРѕР»РёС‡РµСЃС‚РІРѕ РёС‚РµСЂР°С†РёР№ РѕРїС‚РёРјРёР·Р°С†РёРё РЅР° РѕРґРЅСѓ РІРµСЂС€РёРЅСѓ
const IterationsPerNode = 1000

// optimizePositions РѕРїС‚РёРјРёР·РёСЂСѓРµС‚ РїРѕР·РёС†РёРё РІРµСЂС€РёРЅ РјРµС‚РѕРґРѕРј РѕС‚Р¶РёРіР° (simulated annealing)
func optimizePositions(cm *CoordMatrix) {
	// РЎРѕР±РёСЂР°РµРј РІСЃРµ РЅРµ-РїСЃРµРІРґРѕ РІРµСЂС€РёРЅС‹ РІ РґРµС‚РµСЂРјРёРЅРёСЂРѕРІР°РЅРЅРѕРј РїРѕСЂСЏРґРєРµ
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

	// РРЅРёС†РёР°Р»РёР·РёСЂСѓРµРј РіРµРЅРµСЂР°С‚РѕСЂ СЃР»СѓС‡Р°Р№РЅС‹С… С‡РёСЃРµР» СЃ С„РёРєСЃРёСЂРѕРІР°РЅРЅС‹Рј СЃРёРґРѕРј
	rng := rand.New(rand.NewSource(RandomSeed))

	// РљРѕР»РёС‡РµСЃС‚РІРѕ РёС‚РµСЂР°С†РёР№
	iterations := IterationsPerNode * len(allNodes)

	// РџР°СЂР°РјРµС‚СЂС‹ РјРµС‚РѕРґР° РѕС‚Р¶РёРіР°
	initialTemperature := 100.0 // РќР°С‡Р°Р»СЊРЅР°СЏ С‚РµРјРїРµСЂР°С‚СѓСЂР°
	finalTemperature := 0.1     // РљРѕРЅРµС‡РЅР°СЏ С‚РµРјРїРµСЂР°С‚СѓСЂР°

	for iter := 0; iter < iterations; iter++ {
		// Р’С‹С‡РёСЃР»СЏРµРј С‚РµРєСѓС‰СѓСЋ С‚РµРјРїРµСЂР°С‚СѓСЂСѓ (Р»РёРЅРµР№РЅРѕРµ РѕС…Р»Р°Р¶РґРµРЅРёРµ)
		temperature := initialTemperature - (initialTemperature-finalTemperature)*float64(iter)/float64(iterations)

		// Р’С‹Р±РёСЂР°РµРј СЃР»СѓС‡Р°Р№РЅСѓСЋ РІРµСЂС€РёРЅСѓ
		node := allNodes[rng.Intn(len(allNodes))]

		// РџСЂРѕРїСѓСЃРєР°РµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹
		if node.IsPseudo {
			continue
		}

		// РџРѕР»СѓС‡Р°РµРј СЃРѕСЃРµРґРµР№ РЅР° С‚РѕРј Р¶Рµ СЃР»РѕРµ
		leftNeighbor, rightNeighbor := getNeighbors(cm, node)

		// РўРµРєСѓС‰Р°СЏ СЃСѓРјРјР° СЂР°СЃСЃС‚РѕСЏРЅРёР№
		currentDist := calculateTotalDistance(node, cm)

		// РџСЂРѕРІРµСЂСЏРµРј СЃРґРІРёРі РІР»РµРІРѕ
		leftDist := -1
		var leftPushedChain []*CoordNode
		canMoveLeft, pushedLeft := canMoveWithPush(node, -1, leftNeighbor, rightNeighbor, cm)
		if canMoveLeft {
			// РЎРѕС…СЂР°РЅСЏРµРј СЂР°СЃСЃС‚РѕСЏРЅРёСЏ СЃРѕСЃРµРґРµР№ РґРѕ СЃРґРІРёРіР°
			neighborDistsBefore := make([]int, len(pushedLeft))
			for i, neighbor := range pushedLeft {
				neighborDistsBefore[i] = calculateTotalDistance(neighbor, cm)
			}

			// РЎРґРІРёРіР°РµРј
			moveNode(node, -1, cm)
			for _, neighbor := range pushedLeft {
				moveNode(neighbor, -1, cm)
			}
			leftPushedChain = pushedLeft

			// РЎС‡РёС‚Р°РµРј РЅРѕРІРѕРµ СЂР°СЃСЃС‚РѕСЏРЅРёРµ
			leftDist = calculateTotalDistance(node, cm)
			for i, neighbor := range pushedLeft {
				neighborDistAfter := calculateTotalDistance(neighbor, cm)
				leftDist += neighborDistAfter - neighborDistsBefore[i]
			}

			// Р’РѕР·РІСЂР°С‰Р°РµРј РѕР±СЂР°С‚РЅРѕ
			moveNode(node, 1, cm)
			for _, neighbor := range pushedLeft {
				moveNode(neighbor, 1, cm)
			}
		}

		// РџСЂРѕРІРµСЂСЏРµРј СЃРґРІРёРі РІРїСЂР°РІРѕ
		rightDist := -1
		var rightPushedChain []*CoordNode
		canMoveRight, pushedRight := canMoveWithPush(node, 1, leftNeighbor, rightNeighbor, cm)
		if canMoveRight {
			// РЎРѕС…СЂР°РЅСЏРµРј СЂР°СЃСЃС‚РѕСЏРЅРёСЏ СЃРѕСЃРµРґРµР№ РґРѕ СЃРґРІРёРіР°
			neighborDistsBefore := make([]int, len(pushedRight))
			for i, neighbor := range pushedRight {
				neighborDistsBefore[i] = calculateTotalDistance(neighbor, cm)
			}

			// РЎРґРІРёРіР°РµРј
			moveNode(node, 1, cm)
			for _, neighbor := range pushedRight {
				moveNode(neighbor, 1, cm)
			}
			rightPushedChain = pushedRight

			// РЎС‡РёС‚Р°РµРј РЅРѕРІРѕРµ СЂР°СЃСЃС‚РѕСЏРЅРёРµ
			rightDist = calculateTotalDistance(node, cm)
			for i, neighbor := range pushedRight {
				neighborDistAfter := calculateTotalDistance(neighbor, cm)
				rightDist += neighborDistAfter - neighborDistsBefore[i]
			}

			// Р’РѕР·РІСЂР°С‰Р°РµРј РѕР±СЂР°С‚РЅРѕ
			moveNode(node, -1, cm)
			for _, neighbor := range pushedRight {
				moveNode(neighbor, -1, cm)
			}
		}

		// Р’С‹Р±РёСЂР°РµРј Р»СѓС‡С€РёР№ РІР°СЂРёР°РЅС‚ (Р¶Р°РґРЅС‹Р№ РІС‹Р±РѕСЂ)
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

		// РњРµС‚РѕРґ РѕС‚Р¶РёРіР°: РµСЃР»Рё Р»СѓС‡С€РёР№ РІР°СЂРёР°РЅС‚ вЂ” РѕСЃС‚Р°С‚СЊСЃСЏ РЅР° РјРµСЃС‚Рµ,
		// РІСЃС‘ СЂР°РІРЅРѕ СЃ РЅРµРєРѕС‚РѕСЂРѕР№ РІРµСЂРѕСЏС‚РЅРѕСЃС‚СЊСЋ РґРµР»Р°РµРј СЃР»СѓС‡Р°Р№РЅС‹Р№ СЃРґРІРёРі
		if bestMove == 0 && (canMoveLeft || canMoveRight) {
			// Р’С‹Р±РёСЂР°РµРј СЃР»СѓС‡Р°Р№РЅРѕРµ РЅР°РїСЂР°РІР»РµРЅРёРµ РёР· РґРѕСЃС‚СѓРїРЅС‹С…
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

			// Р’С‹Р±РёСЂР°РµРј СЃР»СѓС‡Р°Р№РЅРѕРµ РЅР°РїСЂР°РІР»РµРЅРёРµ
			idx := rng.Intn(len(possibleMoves))
			candidateMove := possibleMoves[idx]
			candidateDist := possibleDists[idx]
			candidateChain := possibleChains[idx]

			// Р’С‹С‡РёСЃР»СЏРµРј РІРµСЂРѕСЏС‚РЅРѕСЃС‚СЊ РїСЂРёРЅСЏС‚РёСЏ С…СѓРґС€РµРіРѕ СЂРµС€РµРЅРёСЏ РїРѕ С„РѕСЂРјСѓР»Рµ Р‘РѕР»СЊС†РјР°РЅР°
			// P = exp(-delta / T), РіРґРµ delta = РЅРѕРІРѕРµ - С‚РµРєСѓС‰РµРµ (РїРѕР»РѕР¶РёС‚РµР»СЊРЅРѕРµ РґР»СЏ СѓС…СѓРґС€РµРЅРёСЏ)
			delta := float64(candidateDist - currentDist)
			if delta <= 0 {
				// Р•СЃР»Рё РЅРµ С…СѓР¶Рµ С‚РµРєСѓС‰РµРіРѕ вЂ” РїСЂРёРЅРёРјР°РµРј
				bestMove = candidateMove
				bestDist = candidateDist
				bestPushedChain = candidateChain
			} else {
				// Р’РµСЂРѕСЏС‚РЅРѕСЃС‚СЊ РїСЂРёРЅСЏС‚РёСЏ С…СѓРґС€РµРіРѕ СЂРµС€РµРЅРёСЏ
				probability := math.Exp(-delta / temperature)
				if rng.Float64() < probability {
					bestMove = candidateMove
					bestDist = candidateDist
					bestPushedChain = candidateChain
				}
			}
		}

		// РћС‚Р»Р°РґРѕС‡РЅС‹Р№ РІС‹РІРѕРґ РґР»СЏ РїРѕСЃР»РµРґРЅРёС… 10 РёС‚РµСЂР°С†РёР№
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

		// РџСЂРёРјРµРЅСЏРµРј Р»СѓС‡С€РёР№ СЃРґРІРёРі
		if bestMove != 0 {
			moveNode(node, bestMove, cm)
			for _, neighbor := range bestPushedChain {
				moveNode(neighbor, bestMove, cm)
			}
		}
	}
}

// getNeighbors РІРѕР·РІСЂР°С‰Р°РµС‚ Р»РµРІРѕРіРѕ Рё РїСЂР°РІРѕРіРѕ СЃРѕСЃРµРґРµР№ РІРµСЂС€РёРЅС‹ РІ СЃР»РѕРµ
func getNeighbors(cm *CoordMatrix, node *CoordNode) (*CoordNode, *CoordNode) {
	var leftNeighbor, rightNeighbor *CoordNode

	nodes := cm.Layers[node.Layer]
	for _, n := range nodes {
		if n == node {
			continue
		}
		// Р›РµРІС‹Р№ СЃРѕСЃРµРґ вЂ” Р±Р»РёР¶Р°Р№С€РёР№ СЃРїСЂР°РІР° РѕС‚ РєРѕС‚РѕСЂРѕРіРѕ РјС‹ РЅР°С…РѕРґРёРјСЃСЏ
		if n.Right <= node.Left {
			if leftNeighbor == nil || n.Right > leftNeighbor.Right {
				leftNeighbor = n
			}
		}
		// РџСЂР°РІС‹Р№ СЃРѕСЃРµРґ вЂ” Р±Р»РёР¶Р°Р№С€РёР№ СЃР»РµРІР° РѕС‚ РєРѕС‚РѕСЂРѕРіРѕ РјС‹ РЅР°С…РѕРґРёРјСЃСЏ
		if n.Left >= node.Right {
			if rightNeighbor == nil || n.Left < rightNeighbor.Left {
				rightNeighbor = n
			}
		}
	}

	return leftNeighbor, rightNeighbor
}

// canMove РїСЂРѕРІРµСЂСЏРµС‚ РјРѕР¶РЅРѕ Р»Рё СЃРґРІРёРЅСѓС‚СЊ РІРµСЂС€РёРЅСѓ РЅР° delta Р±РµР· РїРµСЂРµСЃРµС‡РµРЅРёР№
func canMove(node *CoordNode, delta int, leftNeighbor, rightNeighbor *CoordNode, cm *CoordMatrix) bool {
	newLeft := node.Left + delta
	newRight := node.Right + delta

	// РџСЂРѕРІРµСЂСЏРµРј РїРµСЂРµСЃРµС‡РµРЅРёРµ СЃ СЃРѕСЃРµРґСЏРјРё
	if leftNeighbor != nil && newLeft < leftNeighbor.Right {
		return false
	}
	if rightNeighbor != nil && newRight > rightNeighbor.Left {
		return false
	}

	// Р•СЃР»Рё РµСЃС‚СЊ РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ СЃРІРµСЂС…Сѓ, РїСЂРѕРІРµСЂСЏРµРј РёС… С‚РѕР¶Рµ
	for _, up := range node.Up {
		if up != nil && up.IsPseudo {
			if !canMovePseudoChain(up, delta, cm) {
				return false
			}
		}
	}

	return true
}

// canMoveWithPush РїСЂРѕРІРµСЂСЏРµС‚ РјРѕР¶РЅРѕ Р»Рё СЃРґРІРёРЅСѓС‚СЊ РІРµСЂС€РёРЅСѓ, РІРѕР·РјРѕР¶РЅРѕ С‚РѕР»РєР°СЏ С†РµРїРѕС‡РєСѓ СЃРѕСЃРµРґРµР№
// Р’РѕР·РІСЂР°С‰Р°РµС‚ (РјРѕР¶РЅРѕ Р»Рё СЃРґРІРёРЅСѓС‚СЊ, СЃРїРёСЃРѕРє СЃРѕСЃРµРґРµР№ РєРѕС‚РѕСЂС‹С… РЅСѓР¶РЅРѕ С‚РѕР»РєРЅСѓС‚СЊ)
func canMoveWithPush(node *CoordNode, delta int, leftNeighbor, rightNeighbor *CoordNode, cm *CoordMatrix) (bool, []*CoordNode) {
	// РЎРѕР±РёСЂР°РµРј С†РµРїРѕС‡РєСѓ РІРµСЂС€РёРЅ, РєРѕС‚РѕСЂС‹Рµ РЅСѓР¶РЅРѕ С‚РѕР»РєРЅСѓС‚СЊ
	var pushedChain []*CoordNode
	visited := make(map[*CoordNode]bool)
	visited[node] = true

	// Р РµРєСѓСЂСЃРёРІРЅР°СЏ РїСЂРѕРІРµСЂРєР° С†РµРїРѕС‡РєРё
	canPush := canPushChain(node, delta, leftNeighbor, rightNeighbor, cm, &pushedChain, visited)
	if !canPush {
		return false, nil
	}

	// РџСЂРѕРІРµСЂСЏРµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ СЃРІРµСЂС…Сѓ РґР»СЏ РѕСЃРЅРѕРІРЅРѕР№ РІРµСЂС€РёРЅС‹
	for _, up := range node.Up {
		if up != nil && up.IsPseudo {
			if !canMovePseudoChain(up, delta, cm) {
				return false, nil
			}
		}
	}

	return true, pushedChain
}

// canPushChain СЂРµРєСѓСЂСЃРёРІРЅРѕ РїСЂРѕРІРµСЂСЏРµС‚ РјРѕР¶РЅРѕ Р»Рё С‚РѕР»РєРЅСѓС‚СЊ С†РµРїРѕС‡РєСѓ РІРµСЂС€РёРЅ
func canPushChain(node *CoordNode, delta int, leftNeighbor, rightNeighbor *CoordNode, cm *CoordMatrix, pushedChain *[]*CoordNode, visited map[*CoordNode]bool) bool {
	newLeft := node.Left + delta
	newRight := node.Right + delta

	var blockingNeighbor *CoordNode = nil

	// РџСЂРѕРІРµСЂСЏРµРј РїРµСЂРµСЃРµС‡РµРЅРёРµ СЃ СЃРѕСЃРµРґСЏРјРё
	if delta < 0 && leftNeighbor != nil && newLeft < leftNeighbor.Right {
		// РЎС‚РѕР»РєРЅРѕРІРµРЅРёРµ СЃР»РµРІР°
		if leftNeighbor.IsPseudo {
			return false // РџСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ РЅРµ С‚РѕР»РєР°РµРј РєР°СЃРєР°РґРЅРѕ
		}
		blockingNeighbor = leftNeighbor
	}
	if delta > 0 && rightNeighbor != nil && newRight > rightNeighbor.Left {
		// РЎС‚РѕР»РєРЅРѕРІРµРЅРёРµ СЃРїСЂР°РІР°
		if rightNeighbor.IsPseudo {
			return false // РџСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ РЅРµ С‚РѕР»РєР°РµРј РєР°СЃРєР°РґРЅРѕ
		}
		blockingNeighbor = rightNeighbor
	}

	// Р•СЃР»Рё РЅРµС‚ СЃС‚РѕР»РєРЅРѕРІРµРЅРёСЏ вЂ” РјРѕР¶РЅРѕ РґРІРёРіР°С‚СЊСЃСЏ
	if blockingNeighbor == nil {
		return true
	}

	// РџСЂРѕРІРµСЂСЏРµРј РЅР° С†РёРєР»С‹
	if visited[blockingNeighbor] {
		return false
	}
	visited[blockingNeighbor] = true

	// Р”РѕР±Р°РІР»СЏРµРј СЃРѕСЃРµРґР° РІ С†РµРїРѕС‡РєСѓ
	*pushedChain = append(*pushedChain, blockingNeighbor)

	// РџСЂРѕРІРµСЂСЏРµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ СЃРѕСЃРµРґР°
	for _, up := range blockingNeighbor.Up {
		if up != nil && up.IsPseudo {
			if !canMovePseudoChain(up, delta, cm) {
				return false
			}
		}
	}

	// Р РµРєСѓСЂСЃРёРІРЅРѕ РїСЂРѕРІРµСЂСЏРµРј СЃРѕСЃРµРґРµР№ Р±Р»РѕРєРёСЂСѓСЋС‰РµР№ РІРµСЂС€РёРЅС‹
	neighborLeftN, neighborRightN := getNeighbors(cm, blockingNeighbor)
	// РСЃРєР»СЋС‡Р°РµРј СѓР¶Рµ РїРѕСЃРµС‰С‘РЅРЅС‹Рµ РІРµСЂС€РёРЅС‹
	if visited[neighborLeftN] {
		neighborLeftN = nil
	}
	if visited[neighborRightN] {
		neighborRightN = nil
	}

	return canPushChain(blockingNeighbor, delta, neighborLeftN, neighborRightN, cm, pushedChain, visited)
}

// canMovePseudoChain РїСЂРѕРІРµСЂСЏРµС‚ РјРѕР¶РЅРѕ Р»Рё СЃРґРІРёРЅСѓС‚СЊ С†РµРїРѕС‡РєСѓ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ
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

// moveNode СЃРґРІРёРіР°РµС‚ РІРµСЂС€РёРЅСѓ Рё СЃРІСЏР·Р°РЅРЅС‹Рµ РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ РЅР° delta
func moveNode(node *CoordNode, delta int, cm *CoordMatrix) {
	node.Left += delta
	node.Right += delta

	// РЎРґРІРёРіР°РµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ СЃРІРµСЂС…Сѓ
	for _, up := range node.Up {
		if up != nil && up.IsPseudo {
			movePseudoChainByDelta(up, delta)
		}
	}
}

// movePseudoChainByDelta СЃРґРІРёРіР°РµС‚ С†РµРїРѕС‡РєСѓ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ РЅР° delta
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
