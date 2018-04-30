package game_mgr

type GameCell struct {
	Uuid string
	Host string
	Port string
	Ccu  int32
	CcuMax int32
	ActiveAt int64
}

type SceneCell struct {
	Uuid string
	GameAppId string
	SceneType string
	SceneConfigId string
	Ccu int
	CcuMax int
	ServedServers []string
}

type SceneConf struct {
	ConfId string
	SceneType string
	CcuMax int
}

const (
	SERVER_STATUS_WORKING = iota
	SERVER_STATUS_MAINTAIN
)

const DISPATCH_SERVER = "GameAppDispatcher"

