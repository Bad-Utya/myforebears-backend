package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// AddParents РґРѕР±Р°РІР»СЏРµС‚ СЂРѕРґРёС‚РµР»РµР№ С‡РµР»РѕРІРµРєР° РІ РґРµСЂРµРІРѕ, Р° С‚Р°РєР¶Рµ Р±СЂР°С‚СЊРµРІ Рё СЃРµСЃС‚С‘СЂ
// Р’РѕР·РІСЂР°С‰Р°РµС‚ true РµСЃР»Рё СЂРѕРґРёС‚РµР»Рё Р±С‹Р»Рё РґРѕР±Р°РІР»РµРЅС‹ РёР»Рё РёС… РЅРµС‚
func AddParents(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) bool {
	// Р•СЃР»Рё СЂРѕРґРёС‚РµР»РµР№ РЅРµС‚ вЂ” РЅРёС‡РµРіРѕ РЅРµ РґРµР»Р°РµРј
	if !person.HasParents() {
		return true
	}

	// Р•СЃР»Рё СЂРѕРґРёС‚РµР»Рё СѓР¶Рµ СЂР°Р·РјРµС‰РµРЅС‹ вЂ” РїСЂРѕРїСѓСЃРєР°РµРј (Р±СЂР°С‚СЊСЏ/СЃС‘СЃС‚СЂС‹ СѓР¶Рµ РґРѕР»Р¶РЅС‹ Р±С‹С‚СЊ РґРѕР±Р°РІР»РµРЅС‹)
	if person.Mother.IsLayouted() && person.Father.IsLayouted() {
		return true
	}

	// РџСЂРѕРІРµСЂСЏРµРј РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РЅР° РІС‹СЃРѕС‚Сѓ
	parentLayer := person.Layout.Layer + 1

	// РџСЂРѕРІРµСЂСЏРµРј Р»РµРІРѕРµ РѕРіСЂР°РЅРёС‡РµРЅРёРµ
	if person.Layout.LeftHeightConstraint != nil {
		canAdd, causedBy := person.Layout.LeftHeightConstraint.CanAddAbove(person.Layout.Layer)
		if !canAdd && causedBy != nil {
			// РќСѓР¶РЅРѕ РѕРїСѓСЃС‚РёС‚СЊ РїРѕРґРґРµСЂРµРІРѕ
			visited := NewVisitedSet()
			LowerSubtree(causedBy, true, visited)
			// РџРѕСЃР»Рµ РѕРїСѓСЃРєР°РЅРёСЏ РїСЂРѕР±СѓРµРј СЃРЅРѕРІР°
			return false
		}
	}

	// РџСЂРѕРІРµСЂСЏРµРј РїСЂР°РІРѕРµ РѕРіСЂР°РЅРёС‡РµРЅРёРµ
	if person.Layout.RightHeightConstraint != nil {
		canAdd, causedBy := person.Layout.RightHeightConstraint.CanAddAbove(person.Layout.Layer)
		if !canAdd && causedBy != nil {
			// РќСѓР¶РЅРѕ РѕРїСѓСЃС‚РёС‚СЊ РїРѕРґРґРµСЂРµРІРѕ
			visited := NewVisitedSet()
			LowerSubtree(causedBy, true, visited)
			return false
		}
	}

	// РћРїСЂРµРґРµР»СЏРµРј РїРѕСЂСЏРґРѕРє РґРѕР±Р°РІР»РµРЅРёСЏ СЂРѕРґРёС‚РµР»РµР№ РЅР° РѕСЃРЅРѕРІРµ DirectionConstraint
	var firstParent, secondParent *stage1_input.Person
	var firstDirection, secondDirection stage1_input.PlacementDirection

	switch person.Layout.DirectionConstraint {
	case stage1_input.OnlyRight:
		// РЎРЅР°С‡Р°Р»Р° РїР°РїР° СЃРїСЂР°РІР°, РїРѕС‚РѕРј РјР°РјР° СЃРїСЂР°РІР°
		firstParent = person.Father
		secondParent = person.Mother
		firstDirection = stage1_input.PlacedRight
		secondDirection = stage1_input.PlacedRight
	case stage1_input.OnlyLeft:
		// РЎРЅР°С‡Р°Р»Р° РјР°РјР° СЃР»РµРІР°, РїРѕС‚РѕРј РїР°РїР° СЃР»РµРІР°
		firstParent = person.Mother
		secondParent = person.Father
		firstDirection = stage1_input.PlacedLeft
		secondDirection = stage1_input.PlacedLeft
	default:
		// РџР°РїР° СЃР»РµРІР°, РјР°РјР° СЃРїСЂР°РІР°
		firstParent = person.Father
		secondParent = person.Mother
		firstDirection = stage1_input.PlacedLeft
		secondDirection = stage1_input.PlacedRight
	}

	// Р”РѕР±Р°РІР»СЏРµРј РѕР±РѕРёС… СЂРѕРґРёС‚РµР»РµР№ РєР°Рє РїР°СЂСѓ
	bothParentsNew := !firstParent.IsLayouted() && !secondParent.IsLayouted()

	if bothParentsNew {
		// РћР±Р° СЂРѕРґРёС‚РµР»СЏ РЅРѕРІС‹Рµ вЂ” РґРѕР±Р°РІР»СЏРµРј РєР°Рє РїР°СЂСѓ РІ РѕРґРЅСѓ РІРµСЂС€РёРЅСѓ
		firstParent.Layout = stage1_input.NewPersonLayout(parentLayer)
		firstParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
		firstParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
		firstParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
		queue.Enqueue(firstParent)

		secondParent.Layout = stage1_input.NewPersonLayout(parentLayer)
		secondParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
		secondParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
		secondParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
		queue.Enqueue(secondParent)

		// Р—Р°РїРёСЃС‹РІР°РµРј РІ РёСЃС‚РѕСЂРёСЋ РєР°Рє РїР°СЂСѓ
		history.AddPair(person, firstParent, secondParent, 1, firstDirection, stage1_input.RelationParent)
	} else {
		// РҐРѕС‚СЏ Р±С‹ РѕРґРёРЅ СЂРѕРґРёС‚РµР»СЊ СѓР¶Рµ СЂР°Р·РјРµС‰С‘РЅ вЂ” РґРѕР±Р°РІР»СЏРµРј С‚РѕР»СЊРєРѕ РЅРѕРІРѕРіРѕ
		if !firstParent.IsLayouted() {
			firstParent.Layout = stage1_input.NewPersonLayout(parentLayer)
			firstParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
			firstParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
			firstParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
			queue.Enqueue(firstParent)
			history.Add(person, firstParent, 1, firstDirection, stage1_input.RelationParent)
		}

		if !secondParent.IsLayouted() {
			secondParent.Layout = stage1_input.NewPersonLayout(parentLayer)
			secondParent.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
			secondParent.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
			secondParent.Layout.DirectionConstraint = person.Layout.DirectionConstraint
			queue.Enqueue(secondParent)
			history.Add(person, secondParent, 1, secondDirection, stage1_input.RelationParent)
		}
	}

	// Р”РѕР±Р°РІР»СЏРµРј Р±СЂР°С‚СЊРµРІ Рё СЃРµСЃС‚С‘СЂ РѕС‚ СЌС‚РёС… Р¶Рµ СЂРѕРґРёС‚РµР»РµР№
	// Р—Р°РїРёСЃС‹РІР°РµРј РєР°Рє РґРѕР±Р°РІР»РµРЅРЅС‹Рµ РѕС‚ СЂРѕРґРёС‚РµР»СЏ СЃ РЅСѓР¶РЅРѕР№ СЃС‚РѕСЂРѕРЅС‹
	addSiblings(person, firstParent, secondParent, queue, history)

	return true
}

// addSiblings РґРѕР±Р°РІР»СЏРµС‚ Р±СЂР°С‚СЊРµРІ Рё СЃРµСЃС‚С‘СЂ С‚РµРєСѓС‰РµР№ РІРµСЂС€РёРЅС‹
// leftParentArg вЂ” Р»РµРІС‹Р№ СЂРѕРґРёС‚РµР»СЊ, rightParentArg вЂ” РїСЂР°РІС‹Р№ СЂРѕРґРёС‚РµР»СЊ
func addSiblings(person *stage1_input.Person, leftParentArg *stage1_input.Person, rightParentArg *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	// РџРѕР»СѓС‡Р°РµРј РІСЃРµС… РѕР±С‰РёС… РґРµС‚РµР№ СЂРѕРґРёС‚РµР»РµР№
	siblings := GetCommonChildren(person.Mother, person.Father)

	if len(siblings) <= 1 {
		// РўРѕР»СЊРєРѕ СЃР°Рј person, Р±СЂР°С‚СЊРµРІ/СЃРµСЃС‚С‘СЂ РЅРµС‚
		return
	}

	// РћРїСЂРµРґРµР»СЏРµРј РЅР°РїСЂР°РІР»РµРЅРёРµ РґРѕР±Р°РІР»РµРЅРёСЏ:
	// - OnlyRight в†’ РґРѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР°, С‚РµРєСѓС‰Р°СЏ РІРµСЂС€РёРЅР° = РїРµСЂРІС‹Р№ СЂРµР±С‘РЅРѕРє
	// - OnlyLeft в†’ РґРѕР±Р°РІР»СЏРµРј СЃР»РµРІР°, С‚РµРєСѓС‰Р°СЏ РІРµСЂС€РёРЅР° = РїРѕСЃР»РµРґРЅРёР№ СЂРµР±С‘РЅРѕРє
	// - Р”Р»СЏ РЅР°С‡Р°Р»СЊРЅРѕР№ РІРµСЂС€РёРЅС‹ (IsStartPerson):
	//   - РњСѓР¶С‡РёРЅР° в†’ РґРѕР±Р°РІР»СЏРµРј СЃР»РµРІР°, С‚РµРєСѓС‰Р°СЏ РІРµСЂС€РёРЅР° = РїРѕСЃР»РµРґРЅРёР№ СЂРµР±С‘РЅРѕРє
	//   - Р–РµРЅС‰РёРЅР° в†’ РґРѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР°, С‚РµРєСѓС‰Р°СЏ РІРµСЂС€РёРЅР° = РїРµСЂРІС‹Р№ СЂРµР±С‘РЅРѕРє
	// - РРЅР°С‡Рµ СЃРјРѕС‚СЂРёРј AddedFromLeft:
	//   - Р•СЃР»Рё СЃР»РµРІР° в†’ РґРѕР±Р°РІР»СЏРµРј СЃР»РµРІР°, С‚РµРєСѓС‰Р°СЏ РІРµСЂС€РёРЅР° = РїРѕСЃР»РµРґРЅРёР№ СЂРµР±С‘РЅРѕРє
	//   - Р•СЃР»Рё СЃРїСЂР°РІР° в†’ РґРѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР°, С‚РµРєСѓС‰Р°СЏ РІРµСЂС€РёРЅР° = РїРµСЂРІС‹Р№ СЂРµР±С‘РЅРѕРє
	addSiblingsRight := false
	personIsFirst := false

	switch person.Layout.DirectionConstraint {
	case stage1_input.OnlyRight:
		addSiblingsRight = true
		personIsFirst = true
	case stage1_input.OnlyLeft:
		addSiblingsRight = false
		personIsFirst = false
	default:
		if person.Layout.IsStartPerson {
			// Р”Р»СЏ РЅР°С‡Р°Р»СЊРЅРѕР№ РІРµСЂС€РёРЅС‹: РјСѓР¶С‡РёРЅР° - СЃР»РµРІР°, Р¶РµРЅС‰РёРЅР° - СЃРїСЂР°РІР°
			if person.Gender == stage1_input.Male {
				addSiblingsRight = false
				personIsFirst = false
			} else {
				addSiblingsRight = true
				personIsFirst = true
			}
		} else if person.Layout.AddedFromLeft {
			// Р”РѕР±Р°РІР»РµРЅ СЃР»РµРІР° - СЃРёР±Р»РёРЅРіРё СЃР»РµРІР°
			addSiblingsRight = false
			personIsFirst = false
		} else if person.Layout.DirectionConstraint != stage1_input.NoDirectionConstraint {
			// Р”РѕР±Р°РІР»РµРЅ СЃРїСЂР°РІР° (РЅРµ СЃР»РµРІР° Рё РµСЃС‚СЊ РѕРіСЂР°РЅРёС‡РµРЅРёРµ) - СЃРёР±Р»РёРЅРіРё СЃРїСЂР°РІР°
			addSiblingsRight = true
			personIsFirst = true
		} else {
			// РќРµС‚ РѕРіСЂР°РЅРёС‡РµРЅРёР№ - РїРѕ РїРѕР»Сѓ: РјСѓР¶С‡РёРЅС‹ СЃР»РµРІР°, Р¶РµРЅС‰РёРЅС‹ СЃРїСЂР°РІР°
			if person.Gender == stage1_input.Male {
				addSiblingsRight = false
				personIsFirst = false
			} else {
				addSiblingsRight = true
				personIsFirst = true
			}
		}
	}

	// РћРїСЂРµРґРµР»СЏРµРј Р»РµРІРѕРіРѕ Рё РїСЂР°РІРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ РґР»СЏ РѕРіСЂР°РЅРёС‡РµРЅРёР№ РІС‹СЃРѕС‚С‹
	var leftParent, rightParent *stage1_input.Person
	if person.Father.Layout != nil && person.Mother.Layout != nil {
		// РџСЂРµРґРїРѕР»Р°РіР°РµРј СЃС‚Р°РЅРґР°СЂС‚РЅС‹Р№ РїРѕСЂСЏРґРѕРє: РѕС‚РµС† СЃР»РµРІР°, РјР°С‚СЊ СЃРїСЂР°РІР°
		// РќРѕ СЌС‚Рѕ Р·Р°РІРёСЃРёС‚ РѕС‚ DirectionConstraint
		if person.Layout.DirectionConstraint == stage1_input.OnlyRight {
			leftParent = person.Mother
			rightParent = person.Father
		} else {
			leftParent = person.Father
			rightParent = person.Mother
		}
	} else {
		leftParent = person.Father
		rightParent = person.Mother
	}

	// Р¤РѕСЂРјРёСЂСѓРµРј СЃРїРёСЃРѕРє РґРµС‚РµР№ РІ РїСЂР°РІРёР»СЊРЅРѕРј РїРѕСЂСЏРґРєРµ
	// Р•СЃР»Рё personIsFirst, С‚Рѕ person РїРµСЂРІС‹Р№, РѕСЃС‚Р°Р»СЊРЅС‹Рµ РїРѕСЃР»Рµ
	// Р•СЃР»Рё !personIsFirst, С‚Рѕ СЃРЅР°С‡Р°Р»Р° РѕСЃС‚Р°Р»СЊРЅС‹Рµ, РїРѕС‚РѕРј person
	var orderedChildren []*stage1_input.Person
	if personIsFirst {
		orderedChildren = append(orderedChildren, person)
		for _, sibling := range siblings {
			if sibling.ID != person.ID {
				orderedChildren = append(orderedChildren, sibling)
			}
		}
	} else {
		for _, sibling := range siblings {
			if sibling.ID != person.ID {
				orderedChildren = append(orderedChildren, sibling)
			}
		}
		orderedChildren = append(orderedChildren, person)
	}

	// РћРїСЂРµРґРµР»СЏРµРј РїРѕСЂСЏРґРѕРє РёС‚РµСЂР°С†РёРё РґР»СЏ Р·Р°РїРёСЃРё РІ РёСЃС‚РѕСЂРёСЋ:
	// - РџСЂРё РґРѕР±Р°РІР»РµРЅРёРё СЃРїСЂР°РІР°: РїСЂСЏРјРѕР№ РїРѕСЂСЏРґРѕРє (0, 1, 2, ...)
	// - РџСЂРё РґРѕР±Р°РІР»РµРЅРёРё СЃР»РµРІР°: РѕР±СЂР°С‚РЅС‹Р№ РїРѕСЂСЏРґРѕРє (..., 2, 1, 0)
	var iterOrder []int
	if addSiblingsRight {
		for i := 0; i < len(orderedChildren); i++ {
			iterOrder = append(iterOrder, i)
		}
	} else {
		for i := len(orderedChildren) - 1; i >= 0; i-- {
			iterOrder = append(iterOrder, i)
		}
	}

	// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РІС‹СЃРѕС‚С‹ РґР»СЏ РІСЃРµС… РґРµС‚РµР№
	for _, idx := range iterOrder {
		child := orderedChildren[idx]
		isFirst := (idx == 0)
		isLast := (idx == len(orderedChildren)-1)

		// Р”Р»СЏ С‚РµРєСѓС‰РµР№ РІРµСЂС€РёРЅС‹ (person) РѕР±РЅРѕРІР»СЏРµРј РѕРіСЂР°РЅРёС‡РµРЅРёСЏ
		// Р”Р»СЏ РЅРѕРІС‹С… Р±СЂР°С‚СЊРµРІ/СЃРµСЃС‚С‘СЂ СЃРѕР·РґР°С‘Рј layout
		if child.ID == person.ID {
			// РћР±РЅРѕРІР»СЏРµРј РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РґР»СЏ person
			updateChildHeightConstraints(child, leftParent, rightParent, isFirst, isLast)
		} else {
			// Р•СЃР»Рё Р±СЂР°С‚/СЃРµСЃС‚СЂР° СѓР¶Рµ СЂР°Р·РјРµС‰С‘РЅ вЂ” РїСЂРѕРїСѓСЃРєР°РµРј
			if child.IsLayouted() {
				continue
			}

			child.Layout = stage1_input.NewPersonLayout(person.Layout.Layer)
			child.Layout.DirectionConstraint = person.Layout.DirectionConstraint

			// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЃ РєР°РєРѕР№ СЃС‚РѕСЂРѕРЅС‹ РґРѕР±Р°РІР»РµРЅ
			child.Layout.AddedFromLeft = !addSiblingsRight

			updateChildHeightConstraints(child, leftParent, rightParent, isFirst, isLast)

			queue.Enqueue(child)

			// Р—Р°РїРёСЃС‹РІР°РµРј РІ РёСЃС‚РѕСЂРёСЋ РѕС‚ СЂРѕРґРёС‚РµР»СЏ СЃ РЅСѓР¶РЅРѕР№ СЃС‚РѕСЂРѕРЅС‹
			dir := stage1_input.PlacedLeft
			var fromParent *stage1_input.Person
			if addSiblingsRight {
				dir = stage1_input.PlacedRight
				fromParent = rightParent // РїСЂР°РІС‹Р№ СЂРѕРґРёС‚РµР»СЊ РґР»СЏ РґРѕР±Р°РІР»РµРЅРёСЏ СЃРїСЂР°РІР°
			} else {
				fromParent = leftParent // Р»РµРІС‹Р№ СЂРѕРґРёС‚РµР»СЊ РґР»СЏ РґРѕР±Р°РІР»РµРЅРёСЏ СЃР»РµРІР°
			}
			history.Add(fromParent, child, -1, dir, stage1_input.RelationChild)
		}
	}
}

// updateChildHeightConstraints СѓСЃС‚Р°РЅР°РІР»РёРІР°РµС‚ РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РІС‹СЃРѕС‚С‹ РґР»СЏ СЂРµР±С‘РЅРєР°
func updateChildHeightConstraints(child *stage1_input.Person, leftParent, rightParent *stage1_input.Person, isFirst, isLast bool) {
	if isFirst && isLast {
		// Р•РґРёРЅСЃС‚РІРµРЅРЅС‹Р№ СЂРµР±С‘РЅРѕРє
		child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
		child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
	} else if isFirst {
		// РџРµСЂРІС‹Р№ СЂРµР±С‘РЅРѕРє
		child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
		child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&leftParent.Layout.Layer, child)
	} else if isLast {
		// РџРѕСЃР»РµРґРЅРёР№ СЂРµР±С‘РЅРѕРє
		child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&rightParent.Layout.Layer, child)
		child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
	} else {
		// РЎСЂРµРґРЅРёР№ СЂРµР±С‘РЅРѕРє
		child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&leftParent.Layout.Layer, child)
		child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&rightParent.Layout.Layer, child)
	}
}
