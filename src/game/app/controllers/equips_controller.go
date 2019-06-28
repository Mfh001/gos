package controllers

import (
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/timertask"
	"time"
)

type EquipsController struct {
}

func (*EquipsController) Load(ctx *player.Player, params *pt.EquipLoadParams) interface{} {
	runAt := time.Now().Add(5 * time.Second).Unix()
	logger.INFO("now: ", time.Now().Unix(), runAt)
	err := timertask.Add("fake_task_id", runAt, ctx.PlayerId, &pt.EquipUnLoadParams{
		PlayerID: "player_id",
		EquipId:  "equip_id",
	})
	if err != nil {
		logger.ERR("add timertask failed: ", err)
	}

	return &pt.EquipLoadResponse{
		PlayerID: "player_id",
		EquipId: "equip_id",
		Level: 10,
	}
}

func (*EquipsController) UnLoad(ctx *player.Player, params *pt.EquipUnLoadParams) interface{} {
	logger.INFO("UnLoad equips")
	return &pt.EquipLoadResponse{
		PlayerID: "player_id",
		EquipId: "equip_id",
		Level: 10,
	}
}
