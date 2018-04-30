package gosconf

import (
	"google.golang.org/grpc"
	"time"
)

const IS_DEBUG = true

const (
	TCP_READ_TIMEOUT = 60 * time.Second
	HEARTBEAT = 5 * time.Second
	SERVICE_DEAD_DURATION = 16
	RPC_REQUEST_TIMEOUT = 5 * time.Second
	AGENT_CCU_MAX = 20000
	GAME_CCU_MAX = 5000
)

/*
Redis for service config data
service config datas: session, connectApp, gameApp, scene
*/
type Redis struct {
	Host string
	Password string
	Db int
}

var REDIS_FOR_SERVICE = &Redis{
	Host: "localhost:6379",
	Password: "",
	Db: 0,
}

// Keys for retrive redis data
const (
	RK_GAME_APP_IDS = "__GAME_APP_IDS__"
	RK_SCENE_IDS = "__SCENE_IDS__"
	RK_SCENE_CONF_IDS = "__SCENE_CONF_IDS__"
	RK_DEFAULT_SERVER_SCENE_CONF_ID = "__DEFAULT_SERVER_SCENE_ID__"
)

/*
HTTP Servers
*/

const (
	AUTH_SERVICE_PORT = "3000"
)

/*
TCP Servers
*/

type TCP struct {
	Network string
	Address string
	Port string
}

var TCP_SERVER_CONNECT_APP = &TCP{
	Network: "tcp",
	Address: ":4000",
	Port: "4000",
}

/*
Rpc Servers
*/
type Rpc struct {
	ListenNet string // must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	ListenAddr string
	ListenPort string
	DialAddress string
	DialOptions []grpc.DialOption
}

var RPC_FOR_CONNECT_APP_MGR = &Rpc{
	ListenNet:"tcp4",
	ListenAddr:":50051",
	DialAddress:"127.0.0.1:50051",
	DialOptions:[]grpc.DialOption{grpc.WithInsecure()},
}

var RPC_FOR_GAME_APP_MGR = &Rpc{
	ListenNet:"tcp4",
	ListenAddr:":50052",
	DialAddress:"127.0.0.1:50052",
	DialOptions:[]grpc.DialOption{grpc.WithInsecure()},
}

// Dial addresses are dynamic dispatched by GameAppMgr
var RPC_FOR_GAME_APP_STREAM = &Rpc{
	ListenNet:"tcp4",
	ListenAddr:"127.0.0.1:50053",
	ListenPort: "50053",
	DialAddress:"",
	DialOptions:[]grpc.DialOption{grpc.WithInsecure()},
}
