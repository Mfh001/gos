package models

import (
	. "app/consts"
	"fmt"
	"gslib/base_model"
)

type EquipModel struct {
	base_model.BaseModel
	Data *Equip
}

func (e *EquipModel) Load(heroId string) {
	fmt.Println("Load equip to: ", heroId)
}
