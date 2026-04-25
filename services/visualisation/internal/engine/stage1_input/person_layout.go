package stage1_input

// DirectionConstraint РѕРїСЂРµРґРµР»СЏРµС‚ РѕРіСЂР°РЅРёС‡РµРЅРёРµ РЅР° РЅР°РїСЂР°РІР»РµРЅРёРµ РґРѕР±Р°РІР»РµРЅРёСЏ СЂРѕРґСЃС‚РІРµРЅРЅРёРєРѕРІ
type DirectionConstraint int

const (
	NoDirectionConstraint DirectionConstraint = iota
	OnlyLeft                                  // Р”РѕР±Р°РІР»СЏС‚СЊ РїР°СЂС‚РЅС‘СЂРѕРІ С‚РѕР»СЊРєРѕ СЃР»РµРІР°, СЂРѕРґРёС‚РµР»РµР№: СЃРЅР°С‡Р°Р»Р° РїР°РїР°, РїРѕС‚РѕРј РјР°РјР°
	OnlyRight                                 // Р”РѕР±Р°РІР»СЏС‚СЊ РїР°СЂС‚РЅС‘СЂРѕРІ С‚РѕР»СЊРєРѕ СЃРїСЂР°РІР°, СЂРѕРґРёС‚РµР»РµР№: СЃРЅР°С‡Р°Р»Р° РјР°РјР°, РїРѕС‚РѕРј РїР°РїР°
)

// HeightConstraint РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РѕРіСЂР°РЅРёС‡РµРЅРёРµ РЅР° РјР°РєСЃРёРјР°Р»СЊРЅСѓСЋ РІС‹СЃРѕС‚Сѓ СЃР»РѕСЏ
type HeightConstraint struct {
	MaxLayer *int    // РЈРєР°Р·Р°С‚РµР»СЊ РЅР° РјР°РєСЃРёРјР°Р»СЊРЅС‹Р№ РЅРѕРјРµСЂ СЃР»РѕСЏ
	CausedBy *Person // Р’РµСЂС€РёРЅР°, РёР·-Р·Р° РєРѕС‚РѕСЂРѕР№ РїРѕСЏРІРёР»РѕСЃСЊ РѕРіСЂР°РЅРёС‡РµРЅРёРµ
}

// PersonLayout СЃРѕРґРµСЂР¶РёС‚ РёРЅС„РѕСЂРјР°С†РёСЋ Рѕ СЂР°СЃРїРѕР»РѕР¶РµРЅРёРё С‡РµР»РѕРІРµРєР° РІ РґРµСЂРµРІРµ
type PersonLayout struct {
	Layer int // РќРѕРјРµСЂ СЃР»РѕСЏ: 0 = РЅР°С‡Р°Р»СЊРЅС‹Р№, +1 РґР»СЏ СЂРѕРґРёС‚РµР»РµР№, -1 РґР»СЏ РґРµС‚РµР№

	// РћРіСЂР°РЅРёС‡РµРЅРёСЏ РЅР° РІС‹СЃРѕС‚Сѓ
	LeftHeightConstraint  *HeightConstraint // РћРіСЂР°РЅРёС‡РµРЅРёРµ СЃР»РµРІР°
	RightHeightConstraint *HeightConstraint // РћРіСЂР°РЅРёС‡РµРЅРёРµ СЃРїСЂР°РІР°

	// РћРіСЂР°РЅРёС‡РµРЅРёРµ РЅР° РЅР°РїСЂР°РІР»РµРЅРёРµ
	DirectionConstraint DirectionConstraint

	// РЎ РєР°РєРѕР№ СЃС‚РѕСЂРѕРЅС‹ Р±С‹Р»Р° РґРѕР±Р°РІР»РµРЅР° СЌС‚Р° РІРµСЂС€РёРЅР° (true = СЃР»РµРІР°, false = СЃРїСЂР°РІР°)
	AddedFromLeft bool

	// Р¤Р»Р°Рі РЅР°С‡Р°Р»СЊРЅРѕР№ РІРµСЂС€РёРЅС‹ (С‚Р°, СЃ РєРѕС‚РѕСЂРѕР№ РЅР°С‡Р°Р»Рё РѕР±С…РѕРґ)
	IsStartPerson bool

	// Р¤Р»Р°Рі РѕР±СЂР°Р±РѕС‚РєРё
	Processed bool
}

// NewPersonLayout СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ layout СЃ Р·Р°РґР°РЅРЅС‹Рј СЃР»РѕРµРј
func NewPersonLayout(layer int) *PersonLayout {
	return &PersonLayout{
		Layer:     layer,
		Processed: false,
	}
}

// CopyHeightConstraint СЃРѕР·РґР°С‘С‚ РєРѕРїРёСЋ РѕРіСЂР°РЅРёС‡РµРЅРёСЏ РІС‹СЃРѕС‚С‹
func CopyHeightConstraint(constraint *HeightConstraint) *HeightConstraint {
	if constraint == nil {
		return nil
	}
	return &HeightConstraint{
		MaxLayer: constraint.MaxLayer,
		CausedBy: constraint.CausedBy,
	}
}

// NewHeightConstraint СЃРѕР·РґР°С‘С‚ РЅРѕРІРѕРµ РѕРіСЂР°РЅРёС‡РµРЅРёРµ РІС‹СЃРѕС‚С‹
func NewHeightConstraint(layerPtr *int, causedBy *Person) *HeightConstraint {
	return &HeightConstraint{
		MaxLayer: layerPtr,
		CausedBy: causedBy,
	}
}

// CanAddAbove РїСЂРѕРІРµСЂСЏРµС‚, РјРѕР¶РЅРѕ Р»Рё РґРѕР±Р°РІРёС‚СЊ СЃР»РѕР№ РІС‹С€Рµ С‚РµРєСѓС‰РµРіРѕ
// Р’РѕР·РІСЂР°С‰Р°РµС‚ true РµСЃР»Рё РјРѕР¶РЅРѕ, false Рё CausedBy РµСЃР»Рё РЅРµР»СЊР·СЏ
func (c *HeightConstraint) CanAddAbove(currentLayer int) (bool, *Person) {
	if c == nil || c.MaxLayer == nil {
		return true, nil
	}
	// Р РѕРґРёС‚РµР»Рё РґРѕР»Р¶РЅС‹ Р±С‹С‚СЊ РЅР° СЃР»РѕРµ currentLayer + 1
	// РћРіСЂР°РЅРёС‡РµРЅРёРµ: MaxLayer - РјР°РєСЃРёРјР°Р»СЊРЅС‹Р№ РґРѕРїСѓСЃС‚РёРјС‹Р№ СЃР»РѕР№ РґР»СЏ СЂРѕРґРёС‚РµР»РµР№
	// Р•СЃР»Рё СЂРѕРґРёС‚РµР»Рё С…РѕС‚СЏС‚ РЅР° СЃР»РѕР№ >= MaxLayer, СЌС‚Рѕ Р·Р°РїСЂРµС‰РµРЅРѕ
	parentLayer := currentLayer + 1
	if parentLayer >= *c.MaxLayer {
		return false, c.CausedBy
	}
	return true, nil
}
