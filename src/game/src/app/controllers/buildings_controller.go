package controllers

import (
	"gen/api/pt"
	"gen/db"
	"gen/gd"
	"goslib/player"
)

type BuildingsController struct{}

func (*BuildingsController) Create(ctx *player.Player, params *pt.BuildingCreateParams) (string, interface{}) {
	// 是否解锁
	var center *db.Building
	centerConfId := gd.GetGlobal().BuildingCenter
	for _, building := range ctx.Data.Buildings {
		if building.ConfId == centerConfId {
			center = building
			break
		}
	}
	buildingConf := gd.BuildingsIns.GetItem(params.ConfId)
	if buildingConf.UnlockLevel > center.Level {
		return pt.PT_Fail, &pt.Fail{Fail: "error_building_not_unlock"}
	}

	// 资源是否充足
	// 建筑数量是否达到上限
}

func (*BuildingsController) Upgrade(ctx *player.Player, params *pt.BuildingUpgradeParams) (string, interface{}) {
}
