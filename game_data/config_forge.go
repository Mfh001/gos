package game_data

// 2
type ConfigForge struct {
	EquipId int
	Health  int
	Attack  int
}

var ConfigForges = make(map[int]ConfigForge)

func InitForges() {
	ConfigForges[1] = ConfigForge{1, 2, 3}
	ConfigForges[2] = ConfigForge{2, 3, 4}
}

func FindForge(key int) ConfigForge {
	return ConfigForges[key]
}
