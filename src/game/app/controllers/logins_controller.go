package controllers

import (
	"github.com/mafei198/gos/game/app/models/equip_model"
	"github.com/mafei198/gos/game/app/models/user_model"
	"github.com/mafei198/gos/game/app/notice"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/player"
)

type LoginsController struct{}

func (*LoginsController) Heartbeat(ctx *player.Player, params *pt.HeartbeatParams) interface{} {
	return notice.OK()
}

func (*LoginsController) Login(ctx *player.Player, params *pt.LoginParams) interface{} {
	if err := user_model.CheckInit(ctx); err != nil {
		return notice.Fail(notice.INIT_USER_FAILED)
	}

	return &pt.LoginRsp{
		User:          user_model.Info(ctx),
		Equips:        equip_model.Infos(ctx),
	}
}
