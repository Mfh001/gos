package api

import "utils/packet"

type Decoder struct {
}

func (self *Decoder) DecodeEquipsUnloadParams(data *packet.Packet) (player_id, equip_id, hero_id string) {
	return "playerid", "equipsid", "heroid"
}
