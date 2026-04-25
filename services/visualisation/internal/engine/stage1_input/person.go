package stage1_input

// Gender РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РїРѕР» С‡РµР»РѕРІРµРєР°
type Gender int

const (
	Male Gender = iota
	Female
)

// Person РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРѕСЃР»РѕРІРЅРѕРј РґРµСЂРµРІРµ
type Person struct {
	ID     int
	Name   string
	Gender Gender

	// Р РѕРґСЃС‚РІРµРЅРЅС‹Рµ СЃРІСЏР·Рё
	Mother   *Person   // РњР°С‚СЊ (nil РµСЃР»Рё РЅРµ СѓРєР°Р·Р°РЅР°)
	Father   *Person   // РћС‚РµС† (nil РµСЃР»Рё РЅРµ СѓРєР°Р·Р°РЅ)
	Partners []*Person // Р’СЃРµ РїР°СЂС‚РЅС‘СЂС‹ (РїРµСЂРІС‹Р№ = С‚РµРєСѓС‰РёР№, РѕСЃС‚Р°Р»СЊРЅС‹Рµ = Р±С‹РІС€РёРµ)
	Children []*Person // Р’СЃРµ РґРµС‚Рё (РІ РїРѕСЂСЏРґРєРµ РґРѕР±Р°РІР»РµРЅРёСЏ)

	// РРЅС„РѕСЂРјР°С†РёСЏ Рѕ СЂР°СЃРїРѕР»РѕР¶РµРЅРёРё (Р·Р°РїРѕР»РЅСЏРµС‚СЃСЏ Р°Р»РіРѕСЂРёС‚РјРѕРј)
	Layout *PersonLayout
}

// NewPerson СЃРѕР·РґР°С‘С‚ РЅРѕРІРѕРіРѕ С‡РµР»РѕРІРµРєР°
func NewPerson(id int, name string, gender Gender) *Person {
	return &Person{
		ID:       id,
		Name:     name,
		Gender:   gender,
		Partners: make([]*Person, 0),
		Children: make([]*Person, 0),
	}
}

// HasParents РїСЂРѕРІРµСЂСЏРµС‚, СѓРєР°Р·Р°РЅС‹ Р»Рё РѕР±Р° СЂРѕРґРёС‚РµР»СЏ
func (p *Person) HasParents() bool {
	return p.Mother != nil && p.Father != nil
}

// IsLayouted РїСЂРѕРІРµСЂСЏРµС‚, СЂР°Р·РјРµС‰С‘РЅ Р»Рё С‡РµР»РѕРІРµРє РІ РґРµСЂРµРІРµ
func (p *Person) IsLayouted() bool {
	return p.Layout != nil
}

// IsProcessed РїСЂРѕРІРµСЂСЏРµС‚, РѕР±СЂР°Р±РѕС‚Р°РЅ Р»Рё С‡РµР»РѕРІРµРє Р°Р»РіРѕСЂРёС‚РјРѕРј
func (p *Person) IsProcessed() bool {
	return p.Layout != nil && p.Layout.Processed
}
