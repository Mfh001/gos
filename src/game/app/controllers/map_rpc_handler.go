package controllers

import (
	"github.com/mafei198/gos/game/app/notice"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/player"
)

type MapRpcHandler struct {}

func (*MapRpcHandler) RpcInfo(mapCtx *player.Player, params *pt.MapRpcInfosParams) interface{} {
	return notice.OK()
}