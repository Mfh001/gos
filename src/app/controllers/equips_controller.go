package controllers

import (
	. "app/consts"
	// "fmt"
	"gslib"
)

type EquipsController struct {
	Context *gslib.Player
}

func (self *EquipsController) Load(params *EquipLoadParams) interface{} {
	// fmt.Println("Context: ", self.Context)
	// fmt.Println("SystemInfo: ", self.Context.SystemInfo())
	return &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}
