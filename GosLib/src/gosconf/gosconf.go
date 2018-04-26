package gosconf

import "google.golang.org/grpc"

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
	"localhost:6379",
	"",
	0,
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
}

var TCP_SERVER_CONNECT_APP = &TCP{
	"tcp",
	":4000",
}

/*
Rpc Servers
*/
type Rpc struct {
	ListenNet string // must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	ListenAddr string
	DialAddress string
	DialOptions []grpc.DialOption
}

var RPC_FOR_CONNECT_APP_MGR = &Rpc{
	"tcp4",
	":50051",
	"localhost:50051",
	[]grpc.DialOption{grpc.WithInsecure()},
}

var RPC_FOR_GAME_APP_MGR = &Rpc{
	"tcp4",
	":50052",
	"localhost:50052",
	[]grpc.DialOption{grpc.WithInsecure()},
}

// Dial addresses are dynamic dispatched by GameAppMgr
var RPC_FOR_GAME_APP_STREAM = &Rpc{
	"tcp4",
	"localhost:50053",
	"",
	[]grpc.DialOption{grpc.WithInsecure()},
}
