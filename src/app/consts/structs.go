package consts

// type EquipLoadParams struct {
// 	PlayerID string
// 	EquipId  string
// 	HeroId   string
// }

// type EquipLoadResponse struct {
// 	PlayerID string
// 	EquipId  string
// 	Level    int
// }

type Equip struct {
	Uuid    string `db:"uuid"`
	UserId  string `db:"user_id"`
	Level   int    `db:"level"`
	ConfId  int    `db:"conf_id"`
	Evolves string `db:"evolves"`
	Equiped string `db:"equiped"`
	Exp     int    `db:"exp"`
}
