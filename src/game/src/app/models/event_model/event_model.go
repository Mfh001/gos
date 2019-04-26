package event_model

import (
	"gen/api/pt"
	"gen/db"
	"github.com/rs/xid"
	"goslib/player"
	"goslib/timertask"
	"time"
)

func FindUpgrade(ctx *player.Player, eventId string) *db.UpgradeEvent {
	return ctx.Data.UpgradeEvents[eventId]
}

func DelUpgrade(ctx *player.Player, eventId string) {
	delete(ctx.Data.UpgradeEvents, eventId)
}

func CreateUpgrade(ctx *player.Player, category, targetId, upgradeTo, duration int32) *db.UpgradeEvent {
	uuid := xid.New().String()
	now := time.Now().Unix()
	event := &db.UpgradeEvent{
		Uuid:      uuid,
		UserId:    ctx.PlayerId,
		Category:  category,
		TargetId:  targetId,
		UpgradeTo: upgradeTo,
		CreatedAt: now,
		FinishAt:  now + int64(duration),
		Duration:  duration,
	}
	ctx.Data.UpgradeEvents[uuid] = event
	timertask.Add(uuid, event.FinishAt, ctx.PlayerId, pt.PT_BuildingFinishParams, pt.BuildingFinishParams{
		EventId:    uuid,
		BuildingId: targetId,
	})
	return event
}

func UpgradeInfo(event *db.UpgradeEvent) *pt.UpgradeEvent {
	return &pt.UpgradeEvent{
		Uuid:      event.Uuid,
		Category:  event.Category,
		TargetId:  event.TargetId,
		CreatedAt: event.CreatedAt,
		FinishAt:  event.FinishAt,
		Duration:  event.Duration,
	}
}
