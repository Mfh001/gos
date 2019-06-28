package equip_model

import (
	"github.com/mafei198/gos/game/app/models/data"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/gen/db"
	"github.com/mafei198/gos/goslib/player"
)

func Infos(ctx *player.Player) []*pt.Equip {
	equips := data.Equips(ctx)
	equipsInfo := make([]*pt.Equip, len(equips))
	idx := 0
	for _, equip := range equips {
		equipsInfo[idx] = Info(equip)
		idx++
	}
	return equipsInfo
}

func Info(equip *db.Equip) *pt.Equip {
	return &pt.Equip{
		Id:     equip.Id,
		ConfId: equip.ConfId,
	}
}
