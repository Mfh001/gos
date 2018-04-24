package models

import (
	. "app/consts"
	"fmt"
	"gslib/baseModel"
)

type EquipModel struct {
	baseModel.BaseModel
	Data *Equip
}

func (e *EquipModel) Load(heroId string) {
	fmt.Println("Load equip to: ", heroId)
}
