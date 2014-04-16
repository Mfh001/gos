package api

type Decoder struct {
}

func (self *Decoder) DecodeEquipsUnloadParams(data []byte) (player_id, equip_id, hero_id string) {
	return "playerid", "equipsid", "heroid"
}
