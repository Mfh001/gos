package controllers

import (
	. "api"
	"app/consts"
	"app/models"
	"gslib/player"
)

type EquipsController struct {
	Ctx *player.Player
}

func (self *EquipsController) Load(params *EquipLoadParams) (string, interface{}) {
	user := models.CreateUser(self.Ctx, &consts.User{
		Level:  1,
		Exp:    0,
		Online: true,
	})

	user1 := models.FindUser(self.Ctx, user.GetUuid())
	user1.Data.Level = 10
	user1.Save()

	return "EquipLoadResponse", &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}

func (self *EquipsController) UnLoad(params *EquipUnLoadParams) (string, interface{}) {
	return "EquipLoadResponse", &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}
