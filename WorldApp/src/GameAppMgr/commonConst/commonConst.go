package commonConst

type GameCell struct {
	Uuid string
	Name string
	Host string
	Port string
	Ccu  int
	CcuMax int
	Status int
	ServedScenes []string
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

const GAME_APP_IDS_KEY = "__GAME_CELL_IDS_KEY"
const SCENE_CELL_IDS_KEY = "__SCENE_CELL_IDS_KEY"
const SCENE_CONF_IDS_KEY = "__SCENE_CONF_IDS_KEY"
const DEFAULT_SERVER_SCENE_CONF_ID = "__DEFAULT_SERVER_SCENE_ID"

const SERVER = "GameAppDispatcher"

