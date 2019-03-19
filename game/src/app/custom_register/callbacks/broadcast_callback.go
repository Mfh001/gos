package callbacks

import (
	"goslib/broadcast"
	"goslib/player"
)

func RegisterBroadcast() {
	player.BroadcastHandler = func(ctx *player.Player, msg *broadcast.BroadcastMsg) {
		switch msg.Category {
		case "world_chat":
			ctx.SendData(msg.Category, msg.Data)
		}
	}
}
