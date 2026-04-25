package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// AddChildren РґРѕР±Р°РІР»СЏРµС‚ РѕР±С‰РёС… РґРµС‚РµР№ РґРІСѓС… СЂРѕРґРёС‚РµР»РµР№
// mainParent вЂ” СЂРѕРґРёС‚РµР»СЊ, РѕС‚ РєРѕС‚РѕСЂРѕРіРѕ РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ РґРµС‚Рё (РѕР±С‹С‡РЅРѕ С‚РѕС‚, РєС‚Рѕ РІ РѕС‡РµСЂРµРґРё)
// otherParent вЂ” РІС‚РѕСЂРѕР№ СЂРѕРґРёС‚РµР»СЊ (РїР°СЂС‚РЅС‘СЂ)
func AddChildren(mainParent, otherParent *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	children := GetCommonChildren(mainParent, otherParent)

	if len(children) == 0 {
		return
	}

	// РћРїСЂРµРґРµР»СЏРµРј РЅР°РїСЂР°РІР»РµРЅРёРµ РґРѕР±Р°РІР»РµРЅРёСЏ РґРµС‚РµР№:
	// Р”РµС‚Рё РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ РІ С‚Сѓ Р¶Рµ СЃС‚РѕСЂРѕРЅСѓ, РіРґРµ РЅР°С…РѕРґРёС‚СЃСЏ РІС‚РѕСЂРѕР№ СЂРѕРґРёС‚РµР»СЊ
	// Р­С‚Рѕ РіР°СЂР°РЅС‚РёСЂСѓРµС‚, С‡С‚Рѕ РґРµС‚Рё РѕС‚ РїР°СЂС‚РЅС‘СЂР° СЃРїСЂР°РІР° Р±СѓРґСѓС‚ СЃРїСЂР°РІР°,
	// Р° РґРµС‚Рё РѕС‚ РїР°СЂС‚РЅС‘СЂР° СЃР»РµРІР° вЂ” СЃР»РµРІР°
	var childDirection stage1_input.PlacementDirection
	var leftParent, rightParent *stage1_input.Person

	// РћРїСЂРµРґРµР»СЏРµРј РїРѕР·РёС†РёСЋ otherParent РѕС‚РЅРѕСЃРёС‚РµР»СЊРЅРѕ mainParent
	// РСЃРїРѕР»СЊР·СѓРµРј AddedFromLeft РІС‚РѕСЂРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ, РµСЃР»Рё РѕРЅ Р±С‹Р» РґРѕР±Р°РІР»РµРЅ РѕС‚ mainParent
	// РР»Рё РїСЂРѕРІРµСЂСЏРµРј DirectionConstraint
	otherIsLeft := isPartnerOnLeft(mainParent, otherParent)

	if otherIsLeft {
		// Р’С‚РѕСЂРѕР№ СЂРѕРґРёС‚РµР»СЊ СЃР»РµРІР° в†’ РґРµС‚Рё РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ СЃР»РµРІР°
		childDirection = stage1_input.PlacedLeft
		leftParent = otherParent
		rightParent = mainParent
	} else {
		// Р’С‚РѕСЂРѕР№ СЂРѕРґРёС‚РµР»СЊ СЃРїСЂР°РІР° в†’ РґРµС‚Рё РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ СЃРїСЂР°РІР°
		childDirection = stage1_input.PlacedRight
		leftParent = mainParent
		rightParent = otherParent
	}

	childLayer := mainParent.Layout.Layer - 1

	// РћРїСЂРµРґРµР»СЏРµРј РїРѕСЂСЏРґРѕРє РѕР±С…РѕРґР° РґРµС‚РµР№:
	// - РџСЂРё РґРѕР±Р°РІР»РµРЅРёРё СЃРїСЂР°РІР°: СЃ РЅР°С‡Р°Р»Р° РІ РєРѕРЅРµС† (0, 1, 2, ...)
	// - РџСЂРё РґРѕР±Р°РІР»РµРЅРёРё СЃР»РµРІР°: СЃ РєРѕРЅС†Р° РІ РЅР°С‡Р°Р»Рѕ (..., 2, 1, 0)
	// РќРѕ РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РІС‹СЃРѕС‚С‹ РѕРїСЂРµРґРµР»СЏСЋС‚СЃСЏ РїРѕ РїРѕР·РёС†РёРё РІ РѕСЂРёРіРёРЅР°Р»СЊРЅРѕРј СЃРїРёСЃРєРµ
	var iterOrder []int
	if otherIsLeft {
		// Р”РѕР±Р°РІР»СЏРµРј СЃР»РµРІР° вЂ” РѕР±СЂР°С‚РЅС‹Р№ РїРѕСЂСЏРґРѕРє
		for i := len(children) - 1; i >= 0; i-- {
			iterOrder = append(iterOrder, i)
		}
	} else {
		// Р”РѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР° вЂ” РїСЂСЏРјРѕР№ РїРѕСЂСЏРґРѕРє
		for i := 0; i < len(children); i++ {
			iterOrder = append(iterOrder, i)
		}
	}

	for _, i := range iterOrder {
		child := children[i]

		// Р•СЃР»Рё СЂРµР±С‘РЅРѕРє СѓР¶Рµ СЂР°Р·РјРµС‰С‘РЅ вЂ” РїСЂРѕРїСѓСЃРєР°РµРј
		if child.IsLayouted() {
			continue
		}

		child.Layout = stage1_input.NewPersonLayout(childLayer)

		// РќР°РїСЂР°РІР»РµРЅРёРµ РґРѕР±Р°РІР»РµРЅРёСЏ СЂРµР±С‘РЅРєР°
		child.Layout.AddedFromLeft = (childDirection == stage1_input.PlacedLeft)

		// РћРіСЂР°РЅРёС‡РµРЅРёСЏ РѕРїСЂРµРґРµР»СЏСЋС‚СЃСЏ РїРѕ РїРѕР·РёС†РёРё РІ РѕСЂРёРіРёРЅР°Р»СЊРЅРѕРј СЃРїРёСЃРєРµ
		isFirst := (i == 0)
		isLast := (i == len(children)-1)

		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РІС‹СЃРѕС‚С‹
		if isFirst && isLast {
			// Р•РґРёРЅСЃС‚РІРµРЅРЅС‹Р№ СЂРµР±С‘РЅРѕРє
			child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
			child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
		} else if isFirst {
			// РџРµСЂРІС‹Р№ СЂРµР±С‘РЅРѕРє РІ РѕСЂРёРіРёРЅР°Р»СЊРЅРѕРј СЃРїРёСЃРєРµ (РЅРµ РµРґРёРЅСЃС‚РІРµРЅРЅС‹Р№)
			child.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(leftParent.Layout.LeftHeightConstraint)
			child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
		} else if isLast {
			// РџРѕСЃР»РµРґРЅРёР№ СЂРµР±С‘РЅРѕРє РІ РѕСЂРёРіРёРЅР°Р»СЊРЅРѕРј СЃРїРёСЃРєРµ
			child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
			child.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(rightParent.Layout.RightHeightConstraint)
		} else {
			// РЎСЂРµРґРЅРёР№ СЂРµР±С‘РЅРѕРє
			child.Layout.LeftHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
			child.Layout.RightHeightConstraint = stage1_input.NewHeightConstraint(&mainParent.Layout.Layer, child)
		}

		queue.Enqueue(child)

		// Р—Р°РїРёСЃС‹РІР°РµРј РІ РёСЃС‚РѕСЂРёСЋ РѕС‚ РІС‚РѕСЂРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ (РїР°СЂС‚РЅС‘СЂР°)
		history.Add(otherParent, child, -1, childDirection, stage1_input.RelationChild)
	}
}

// isPartnerOnLeft РѕРїСЂРµРґРµР»СЏРµС‚, РЅР°С…РѕРґРёС‚СЃСЏ Р»Рё РїР°СЂС‚РЅС‘СЂ СЃР»РµРІР° РѕС‚ РѕСЃРЅРѕРІРЅРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ
// Рё СЃРѕРѕС‚РІРµС‚СЃС‚РІРµРЅРЅРѕ РєСѓРґР° РґРѕР±Р°РІР»СЏС‚СЊ РґРµС‚РµР№
func isPartnerOnLeft(mainParent, otherParent *stage1_input.Person) bool {
	// РџСЂРѕРІРµСЂСЏРµРј DirectionConstraint РѕСЃРЅРѕРІРЅРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ
	if mainParent.Layout.DirectionConstraint == stage1_input.OnlyLeft {
		// РћРіСЂР°РЅРёС‡РµРЅРёРµ "С‚РѕР»СЊРєРѕ СЃР»РµРІР°" в†’ РґРµС‚Рё РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ СЃР»РµРІР°
		return true
	}
	if mainParent.Layout.DirectionConstraint == stage1_input.OnlyRight {
		// РћРіСЂР°РЅРёС‡РµРЅРёРµ "С‚РѕР»СЊРєРѕ СЃРїСЂР°РІР°" в†’ РґРµС‚Рё РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ СЃРїСЂР°РІР°
		return false
	}

	// РќРµС‚ РѕРіСЂР°РЅРёС‡РµРЅРёР№ в†’ РґРµС‚Рё РґРѕР±Р°РІР»СЏСЋС‚СЃСЏ СЃРїСЂР°РІР° (РїРѕ СѓРјРѕР»С‡Р°РЅРёСЋ)
	return false
}
