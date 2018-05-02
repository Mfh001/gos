package controllers

import (
	. "api"
	"gslib/player"
	"goslib/logger"
)

type EquipsController struct {
	Ctx *player.Player
}

func (self *EquipsController) Load(params *EquipLoadParams) (string, interface{}) {
	//player := self.Ctx
	//player.Store.EnsureDataLoaded("equips", "54BC69792B897814D763403B")
	//equipId := "54BC69792B897814D7634040"
	//equipModel := player.Store.Get([]string{"models", "equips"}, equipId).(*EquipModel)
	//equipModel.Load("hahah")
	//fmt.Println("old level: ", equipModel.Data.Level)
	//
	//equipModel.Data.Level = 0
	//equipModel.Save()
	//
	//player.Store.Persist([]string{"models"})
	//fmt.Println("new level: ", equipModel.Data.Level)
	//
	//player.Store.Set([]string{"models", "equips"}, "testUuid", &Equip{Uuid: "testUuid"})
	//
	//equipModel = player.Store.Get([]string{"models", "equips"}, "testUuid").(*EquipModel)
	//
	//fmt.Println("equipModel: ", equipModel)
	//fmt.Println("new level: ", equipModel.Data.Uuid)
	logger.INFO("EquipsController Load!")
	return "EquipLoadResponse", &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}

func (self *EquipsController) UnLoad(params *EquipUnLoadParams) (string, interface{}) {
	 //fmt.Println("Context: ", self.Context)
	 //fmt.Println("SystemInfo: ", self.Context.SystemInfo())
	return "EquipLoadResponse", &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}
