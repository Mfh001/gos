package user_model

import (
	"goslib/player"
	"goslib/utils"
)

func GemEnough(ctx *player.Player, gem int32) bool {
	return ctx.Data.User.Gem >= gem
}

func UseGem(ctx *player.Player, gem int32) int32 {
	user := ctx.Data.User
	user.Gem -= utils.MinInt32(user.Gem, gem)
	return user.Gem
}
