package controllers

// import "fmt"

type EquipsController struct {
}

type Material struct {
	category string
	uuid     string
	amount   int
}

func (self *EquipsController) Load(player_id, equip_id, hero_id string) (string, string, string, string) {
	// fmt.Printf("player_id: %s, equip_id: %s, hero_id: %s \n", player_id, equip_id, hero_id)
	return "EncodeEquip", player_id, equip_id, hero_id
}

func (self *EquipsController) Unload(player_id, equip_id, hero_id string) {
}

func (self *EquipsController) evolve(player_id, equip_id string, material []Material) {
}
