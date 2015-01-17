package models

import (
	. "app/consts"
	"gslib"
)

type EquipModel struct {
	Ctx  *gslib.Player
	Data *Equip
}

func InitEquip(ctx *gslib.Player, data *Equip) *EquipModel {
	return &EquipModel{ctx, data}
}
