package controllers

import (
	"api"
	"app/consts"
	"app/models"
	"goslib/logger"
	"gslib/player"
	"gslib/timertask"
	"time"
)

type EquipsController struct {
	Ctx *player.Player
}

func (self *EquipsController) Load(params *api.EquipLoadParams) (string, interface{}) {
	user := models.CreateUser(self.Ctx, &consts.User{
		Level:  1,
		Exp:    0,
		Online: true,
	})

	user1 := models.FindUser(self.Ctx, user.GetUuid())
	user1.Data.Level = 10
	user1.Save()

	runAt := time.Now().Add(5 * time.Second).Unix()
	timertask.Add("fake_task_id", runAt, self.Ctx.PlayerId, "EquipUnLoadParams", &api.EquipUnLoadParams{
		PlayerID: "player_id",
		EquipId:  "equip_id",
	})

	return api.PT_EquipLoadResponse, &api.EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}

func (self *EquipsController) UnLoad(params *api.EquipUnLoadParams) (string, interface{}) {
	logger.INFO("UnLoad equips")
	return api.PT_EquipLoadResponse, &api.EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}
