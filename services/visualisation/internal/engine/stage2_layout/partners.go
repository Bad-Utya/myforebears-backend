package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// AddPartners РґРѕР±Р°РІР»СЏРµС‚ РїР°СЂС‚РЅС‘СЂРѕРІ С‡РµР»РѕРІРµРєР° Рё РёС… РѕР±С‰РёС… РґРµС‚РµР№
func AddPartners(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	for _, partner := range person.Partners {
		// РћРїСЂРµРґРµР»СЏРµРј, РєСѓРґР° РґРѕР±Р°РІР»СЏС‚СЊ РїР°СЂС‚РЅС‘СЂР°
		addLeft := ShouldAddPartnerLeft(person, partner, person.Layout.DirectionConstraint)

		// Р•СЃР»Рё РїР°СЂС‚РЅС‘СЂ РµС‰С‘ РЅРµ СЂР°Р·РјРµС‰С‘РЅ вЂ” РґРѕР±Р°РІР»СЏРµРј
		if !partner.IsLayouted() {
			partner.Layout = stage1_input.NewPersonLayout(person.Layout.Layer)

			// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЃ РєР°РєРѕР№ СЃС‚РѕСЂРѕРЅС‹ РґРѕР±Р°РІР»РµРЅ
			partner.Layout.AddedFromLeft = addLeft

			// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РІС‹СЃРѕС‚С‹
			if addLeft {
				// Р”РѕР±Р°РІР»СЏРµРј СЃР»РµРІР°: РѕР±Р° РѕРіСЂР°РЅРёС‡РµРЅРёСЏ = Р»РµРІРѕРјСѓ РѕРіСЂР°РЅРёС‡РµРЅРёСЋ С‚РµРєСѓС‰РµР№ РІРµСЂС€РёРЅС‹
				partner.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
				partner.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.LeftHeightConstraint)
			} else {
				// Р”РѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР°: РѕР±Р° РѕРіСЂР°РЅРёС‡РµРЅРёСЏ = РїСЂР°РІРѕРјСѓ РѕРіСЂР°РЅРёС‡РµРЅРёСЋ С‚РµРєСѓС‰РµР№ РІРµСЂС€РёРЅС‹
				partner.Layout.LeftHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
				partner.Layout.RightHeightConstraint = stage1_input.CopyHeightConstraint(person.Layout.RightHeightConstraint)
			}

			// Р•СЃР»Рё Сѓ С‚РµРєСѓС‰РµР№ РІРµСЂС€РёРЅС‹ РµСЃС‚СЊ СЂРѕРґРёС‚РµР»Рё вЂ” СѓСЃС‚Р°РЅР°РІР»РёРІР°РµРј DirectionConstraint
			if person.HasParents() {
				if addLeft {
					partner.Layout.DirectionConstraint = stage1_input.OnlyLeft
				} else {
					partner.Layout.DirectionConstraint = stage1_input.OnlyRight
				}
			}

			queue.Enqueue(partner)

			// Р—Р°РїРёСЃС‹РІР°РµРј РІ РёСЃС‚РѕСЂРёСЋ
			direction := stage1_input.PlacedRight
			if addLeft {
				direction = stage1_input.PlacedLeft
			}
			history.Add(person, partner, 0, direction, stage1_input.RelationPartner)
		}

		// РџРѕСЃР»Рµ РґРѕР±Р°РІР»РµРЅРёСЏ РїР°СЂС‚РЅС‘СЂР° вЂ” РґРѕР±Р°РІР»СЏРµРј РѕР±С‰РёС… РґРµС‚РµР№
		AddChildren(person, partner, queue, history)
	}
}
