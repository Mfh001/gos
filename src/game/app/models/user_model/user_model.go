package user_model

import (
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/redisdb"
	"strconv"
)

func CheckInit(ctx *player.Player) error {
	data := ctx.Data
	if !data.Inited {
		if err := Init(ctx); err != nil {
			return err
		}
	}

	data.Inited = true
	return nil
}


const GuestUserKey = "__GuestUserKey__"

func Init(ctx *player.Player) error {
	idx, err := redisdb.Instance().Incr(GuestUserKey).Result()
	if err != nil {
		return err
	}
	user := ctx.Data.User
	user.Name = "Guest" + strconv.Itoa(int(idx))
	return err
}

func Info(ctx *player.Player) *pt.User {
	user := ctx.Data.User
	return &pt.User{
		Level:           user.Level,
		Exp:             user.Exp,
		Name:            user.Name,
		MaxMonsterLevel: user.MaxMonsterLevel,
	}
}
