package api

import (
	. "app/consts"
	"gslib/utils/packet"
)

func Decode(decode_method string, data *packet.Packet) interface{} {
	if decode_method == "DecodeEquipsUnloadParams" {
		return &EquipLoadParams{"playerid", "equipsid", "heroid"}
	} else {
		return nil
	}
}
