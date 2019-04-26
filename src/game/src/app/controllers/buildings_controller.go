package controllers

import (
	"app/models/building_model"
	"app/models/event_model"
	"app/models/res_model"
	"app/models/user_model"
	"app/notice"
	"gen/api/pt"
	"gen/db"
	"gen/gd"
	"goslib/player"
)

type BuildingsController struct{}

/*
 * API
 */
func (*BuildingsController) Create(ctx *player.Player, params *pt.BuildingCreateParams) (string, interface{}) {
	// 是否解锁
	center := building_model.GetCenter(ctx)
	buildingConf := gd.BuildingsIns.GetItem(params.ConfId)
	if buildingConf.UnlockLevel > center.Level {
		return notice.Fail(notice.BUILDING_NOT_UNLOCK)
	}

	// 建筑数量
	amount := building_model.Amount(ctx, params.ConfId)
	amountLimit := building_model.AmountLimit(ctx, params.ConfId)
	if amount >= amountLimit {
		return notice.Fail(notice.REACH_MAX_BUILDING_AMOUNT)
	}

	if params.Instant {
		return buyCreate(ctx, params)
	} else {
		return resCreate(ctx, params)
	}
}

func (*BuildingsController) Upgrade(ctx *player.Player, params *pt.BuildingUpgradeParams) (string, interface{}) {
	// 是否存在
	building := building_model.Find(ctx, params.BuildingId)
	if building == nil {
		return notice.Fail(notice.BUILDING_NOT_EXIST)
	}

	// 等级上限
	confBuilding := gd.BuildingsIns.GetItem(building.ConfId)
	if building.Level >= confBuilding.MaxLevel {
		return notice.Fail(notice.BUILDING_REACH_MAX_LEVEL)
	}

	if params.Instant {
		return buyUpgrade(ctx, building)
	} else {
		return resUpgrade(ctx, building)
	}
}

// Called by timertask
func (*BuildingsController) Finish(ctx *player.Player, params *pt.BuildingFinishParams) (string, interface{}) {
	event := event_model.FindUpgrade(ctx, params.EventId)
	if event == nil {
		return notice.OK()
	}

	building := building_model.Find(ctx, params.BuildingId)
	if building == nil {
		return notice.OK()
	}

	event_model.DelUpgrade(ctx, event.Uuid)
	building.Level = event.UpgradeTo

	_ = ctx.SendData(ctx.PlayerId, pt.PT_BuildingFinishRsp, &pt.BuildingFinishRsp{
		EventId:       event.Uuid,
		BuildingId:    building.Id,
		BuildingLevel: building.Level,
	})

	return notice.OK()
}

/*
 * Real Actions
 */
func buyCreate(ctx *player.Player, params *pt.BuildingCreateParams) (string, interface{}) {
	upgradeConf := building_model.GetCreateConf(ctx, params.ConfId)
	fee := buyUpgradeFee(ctx, upgradeConf)

	if !user_model.GemEnough(ctx, fee) {
		return notice.Fail(notice.NOT_ENOUGH_GEMS)
	}

	// 使用资源
	resInfo := res_model.Use(ctx, upgradeConf.ResReq)

	// 使用钻石
	remainGem := user_model.UseGem(ctx, fee)

	// 创建建筑
	building := building_model.Create(ctx, params.ConfId, params.Pos)
	building.Level = 1

	// 返回信息
	return pt.PT_BuildingBuyUpgradeRsp, &pt.BuildingBuyUpgradeRsp{
		Gem:        remainGem,
		ResAmounts: resInfo,
		Building:   building_model.Info(building),
	}
}

func resCreate(ctx *player.Player, params *pt.BuildingCreateParams) (string, interface{}) {
	upgradeConf := building_model.GetCreateConf(ctx, params.ConfId)

	// 升级检测
	if fail, msg := checkQueueAndRes(ctx, upgradeConf); fail != "" {
		return fail, msg
	}

	// 创建建筑
	building := building_model.Create(ctx, params.ConfId, params.Pos)

	// 创建事件
	resInfo, event := createUpgradeEvent(ctx, building, upgradeConf)

	// 返回信息
	return pt.PT_BuildingCreateRsp, &pt.BuildingCreateRsp{
		ResAmounts: resInfo,
		Event:      event_model.UpgradeInfo(event),
		Building:   building_model.Info(building),
	}
}

func buyUpgrade(ctx *player.Player, building *db.Building) (string, interface{}) {
	upgradeConf := building_model.GetUpgradeConf(building)
	fee := buyUpgradeFee(ctx, upgradeConf)

	if !user_model.GemEnough(ctx, fee) {
		return notice.Fail(notice.NOT_ENOUGH_GEMS)
	}

	// 使用资源
	resInfo := res_model.Use(ctx, upgradeConf.ResReq)

	// 使用钻石
	remainGem := user_model.UseGem(ctx, fee)

	// 创建建筑
	building.Level++

	// 返回信息
	return pt.PT_BuildingBuyUpgradeRsp, &pt.BuildingBuyUpgradeRsp{
		Gem:        remainGem,
		ResAmounts: resInfo,
		Building:   building_model.Info(building),
	}
}

func resUpgrade(ctx *player.Player, building *db.Building) (string, interface{}) {
	upgradeConf := building_model.GetUpgradeConf(building)

	// 升级检测
	if fail, msg := checkQueueAndRes(ctx, upgradeConf); fail != "" {
		return fail, msg
	}

	// 创建事件
	resInfo, event := createUpgradeEvent(ctx, building, upgradeConf)

	// 返回信息
	return pt.PT_BuildingUpgradeRsp, &pt.BuildingUpgradeRsp{
		ResAmounts: resInfo,
		Event:      event_model.UpgradeInfo(event),
	}
}

/*
 * Checkers
 */
func checkQueueAndRes(ctx *player.Player, upgradeConf *gd.ConfigBuildingUpgrade) (string, interface{}) {
	// 是否有建筑队列
	if !building_model.HaveQueue(ctx) {
		return notice.Fail(notice.BUILDING_QUEUE_FULL)
	}

	// 资源是否充足
	if !res_model.Enough(ctx, upgradeConf.ResReq) {
		return notice.Fail(notice.NOT_ENOUGH_RESOURCE)
	}

	return "", nil
}

/*
 * Helpers
 */
func createUpgradeEvent(ctx *player.Player, building *db.Building,
	upgradeConf *gd.ConfigBuildingUpgrade) ([]*pt.ResAmount, *db.UpgradeEvent) {

	// 使用资源
	resInfo := res_model.Use(ctx, upgradeConf.ResReq)

	// 创建事件
	event := event_model.CreateUpgrade(
		ctx, gd.GetGlobal().EventTypeBuilding,
		building.Id, building.Level+1, upgradeConf.Duartion)

	return resInfo, event
}

func buyUpgradeFee(ctx *player.Player, upgradeConf *gd.ConfigBuildingUpgrade) int32 {
	fee := res_model.AutoBuyFee(ctx, upgradeConf.ResReq)
	fee += res_model.TimeFee(upgradeConf.Duartion)
	return fee
}
