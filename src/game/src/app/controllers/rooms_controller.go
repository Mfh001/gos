package controllers

import (
	"gen/api/pt"
	"goslib/broadcast"
	"goslib/logger"
	"goslib/player"
)

type RoomsController struct {
}

func (*RoomsController) Join(roomCtx *player.Player, params *pt.RoomJoinParams) (string, interface{}) {
	err := broadcast.Join(params.RoomId, params.PlayerId, func(msg *broadcast.BroadcastMsg) {
		roomCtx.SendData(params.PlayerId, msg.Category, msg.Data)
	})
	if err != nil {
		return pt.PT_Fail, pt.Fail{
			Fail: "join room failed",
		}
	}

	err = broadcast.Publish(params.RoomId, params.PlayerId, pt.PT_RoomJoinNotice, &pt.RoomJoinNotice{
		RoomId:      params.RoomId,
		NewPlayerId: params.PlayerId,
	})
	logger.ERR("broadcast join msg failed: ", err)
	return pt.PT_RoomJoinResponse, &pt.RoomJoinResponse{Success: true}
}
