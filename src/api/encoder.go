package api

import "utils/packet"

type Encoder struct {
}

func (self *Encoder) EncodeEquip(player_id string, equip_id string, hero_id string) *packet.Packet {
	writer := packet.Writer()
	writer.WriteString("hello")
	return writer
}
