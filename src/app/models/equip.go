package models

type Equip struct {
	Uuid   string
	UserId string
	ConfId int
}

var EquipMap = make(map[string]*Equip)

func FindEquip(uuid string) *Equip {
	equip, ok := EquipMap[uuid]
	if ok {
		return equip
	} else {
		return equip
	}
}

func LoadData() {

}
