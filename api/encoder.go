package api

type Encoder struct {
}

func (self *Encoder) EncodeEquip(player_id string, equip_id string, hero_id string) []byte {
	return []byte("hello")
}
