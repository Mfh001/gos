package controllers

import (
	. "app/consts"
	"fmt"
	"gslib"
)

type EquipsController struct {
	Context *gslib.Player
}

// func (self *EquipsController) Load(player *player.Player, params *EquipLoadParams) (struct_name string, struct_instance interface{}) {
// 	if equip, equip_exist := models.Equip.find(params.PlayerID, params.EquipId); !equip_exist {
// 		return "Fail", &Fail{code: 298, msg: "equip_not_exist"}
// 	}

// 	if hero, hero_exist := models.Hero.find(params.PlayerID, params.HeroId); !hero_exist {
// 		return "Fail", &Fail{code: 298, msg: "hero_not_exist"}
// 	}

// 	equip.Equiped = params.EquipId
// 	slot = "armor_slot"

// 	old_equiped = reflect.ValueOf(&hero).Elem().FieldByName(slot).String()
// 	if old_equiped != "" {
// 		old_equip = models.Equip.find(params.PlayerID, old_equiped)
// 		old_equip.equiped = ""
// 		player.SendData("Equip", &old_equip)
// 	}

// 	reflect.ValueOf(&hero).Elem().FieldByName(slot).SetString(params.EquipId)

// 	return "Equip", &equip
// }

func (self *EquipsController) Load(params *EquipLoadParams) interface{} {
	fmt.Println("Context: ", self.Context)
	fmt.Println("SystemInfo: ", self.Context.SystemInfo())
	return &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}
