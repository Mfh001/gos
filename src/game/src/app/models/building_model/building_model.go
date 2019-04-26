package building_model

import (
	"gen/api/pt"
	"gen/db"
	"gen/gd"
	"goslib/player"
	"time"
)

func GetCenter(ctx *player.Player) *db.Building {
	return FindByConfId(ctx, gd.GetGlobal().BuildingCenter)
}

func Find(ctx *player.Player, id int32) *db.Building {
	return ctx.Data.Buildings[id]
}

func FindByConfId(ctx *player.Player, confId int32) *db.Building {
	for _, building := range ctx.Data.Buildings {
		if building.ConfId == confId {
			return building
		}
	}
	return nil
}

func FindByPos(ctx *player.Player, pos int32) *db.Building {
	for _, building := range ctx.Data.Buildings {
		if building.Pos == pos {
			return building
		}
	}
	return nil
}

func GetUpgradeConf(building *db.Building) *gd.ConfigBuildingUpgrade {
	upgradeConfId := gd.GetGlobal().BuildingBaseConfID + building.ConfId*100 + building.Level + 1
	return gd.BuildingUpgradeIns.GetItem(upgradeConfId)
}

func GetCreateConf(ctx *player.Player, confId int32) *gd.ConfigBuildingUpgrade {
	upgradeConfId := gd.GetGlobal().BuildingBaseConfID + confId*100 + 1
	return gd.BuildingUpgradeIns.GetItem(upgradeConfId)
}

func AmountLimit(ctx *player.Player, confId int32) int32 {
	center := GetCenter(ctx)
	conf := gd.BuildingsIns.GetItem(confId)
	size := len(conf.Amounts)
	if size == 0 {
		return 1
	}

	for i := size - 1; i >= 0; i-- {
		item := conf.Amounts[i]
		if center.Level >= item.BaseLevel {
			return item.Amount
		}
	}
	return 0
}

func Amount(ctx *player.Player, confId int32) int32 {
	var amount int32
	for _, building := range ctx.Data.Buildings {
		if building.ConfId == confId {
			amount++
		}
	}
	return amount
}

func HaveQueue(ctx *player.Player) bool {
	now := time.Now().Unix()
	for _, builder := range ctx.Data.Builders {
		if (builder.ExpireAt == -1 || int64(builder.ExpireAt) > now) && !builder.IsWorking {
			return true
		}
	}
	return false
}

func Create(ctx *player.Player, confId int32, pos int32) *db.Building {
	id := int32(len(ctx.Data.Buildings) + 1)
	building := &db.Building{
		Id:     id,
		ConfId: confId,
		Pos:    pos,
		Level:  0,
	}
	ctx.Data.Buildings[id] = building
	return building
}

func Info(building *db.Building) *pt.Building {
	return &pt.Building{
		Id:     building.Id,
		ConfId: building.ConfId,
		Pos:    building.Pos,
		Level:  building.Level,
	}
}
