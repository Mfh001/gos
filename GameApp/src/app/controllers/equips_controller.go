package controllers

import (
	. "app/consts"
	 . "app/models"
	"gslib"
	"fmt"
)

type EquipsController struct {
	Ctx *gslib.Player
}

func (self *EquipsController) Load(params *EquipLoadParams) (string, interface{}) {
	player := self.Ctx
	player.Store.LoadData("equips", "54BC69792B897814D763403B")
	equipId := "54BC69792B897814D7634040"
	//equipModel := player.Store.Get([]string{"models", "equips"}, equipId).(*EquipModel)
	equipModel := player.Store.Get("equips", equipId).(*EquipModel)
	equipModel.Load("hahah")
	fmt.Println("old level: ", equipModel.Data.Level)

	equipModel.Data.Level = 0
	equipModel.Save()

	player.Store.Persist([]string{"models"})
	fmt.Println("new level: ", equipModel.Data.Level)

	gslib.CreateModel(self.Ctx, &EquipModel{Data: &Equip{Uuid: "test udid"}})

	equipModel = gslib.FindModel(self.Ctx, "equips", "test udid").(*EquipModel)

	fmt.Println("equipModel: ", equipModel)
	fmt.Println("new level: ", equipModel.Data.Uuid)

	 return "EquipLoadResponse", &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}

func (self *EquipsController) UnLoad(params *EquipUnLoadParams) (string, interface{}) {
	 //fmt.Println("Context: ", self.Context)
	 //fmt.Println("SystemInfo: ", self.Context.SystemInfo())
	return "EquipLoadResponse", &EquipLoadResponse{PlayerID: "player_id", EquipId: "equip_id", Level: 10}
}
