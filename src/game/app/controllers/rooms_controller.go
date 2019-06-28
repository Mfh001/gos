package controllers

import (
	"github.com/mafei198/gos/goslib/broadcast"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/player"
)

type RoomsController struct {
}

func (*RoomsController) Join(roomCtx *player.Player, params *pt.RoomJoinParams) interface{} {
	err := broadcast.Join(params.RoomId, params.PlayerId, func(subscriber string, msg *broadcast.BroadcastMsg) {
		roomCtx.SendData(params.PlayerId, msg.Data)
	})
	if err != nil {
		return pt.Fail{
			Fail: "join room failed",
		}
	}

	err = broadcast.Publish(params.RoomId, params.PlayerId, pt.PT_RoomJoinNotice, &pt.RoomJoinNotice{
		RoomId:      params.RoomId,
		NewPlayerId: params.PlayerId,
	})
	logger.ERR("broadcast join msg failed: ", err)
	return &pt.RoomJoinResponse{Success: true}
}
