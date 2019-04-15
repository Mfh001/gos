package gosconf

import (
	"google.golang.org/grpc"
	"time"
)

const IS_DEBUG = false

const (
	START_TYPE_ALL_IN_ONE = iota
	START_TYPE_CLUSTER
	START_TYPE_K8S
)

const START_TYPE = START_TYPE_ALL_IN_ONE

// supported protocols
const (
	AGENT_PROTOCOL_TCP = iota
	AGENT_PROTOCOL_WS
)

// supported encodings
const (
	PROTOCOL_ENCODING_RAW = iota
	PROTOCOL_ENCODING_JSON
	PROTOCOL_ENCODING_PB
)

const (
	AGENT_PROTOCOL = AGENT_PROTOCOL_TCP
	AGENT_ENCODING = PROTOCOL_ENCODING_PB
)

var MYSQL_DSN_ALL_IN_ONE = "root@/gos_server_development"
var MYSQL_DSN_CLUSTER = "root@/gos_server_development"
var MYSQL_DSN_K8S = "root:euQRdwMgb1@tcp(single-mysql.default.svc.cluster.local)/gos_server_development"

var REDIS_CLUSTERS_FOR_ALL_IN_ONE = []string{
	"127.0.0.1:7000",
	"127.0.0.1:7001",
	"127.0.0.1:7002",
	"127.0.0.1:7003",
	"127.0.0.1:7004",
	"127.0.0.1:7005",
}

var REDIS_CLUSTERS_FOR_CLUSTER = []string{
	"127.0.0.1:7000",
	"127.0.0.1:7001",
	"127.0.0.1:7002",
	"127.0.0.1:7003",
	"127.0.0.1:7004",
	"127.0.0.1:7005",
}

var REDIS_CLUSTERS_FOR_K8S = []string{
	"redis-cluster-0.redis-cluster.default.svc.cluster.local:6379",
	"redis-cluster-1.redis-cluster.default.svc.cluster.local:6379",
	"redis-cluster-2.redis-cluster.default.svc.cluster.local:6379",
	"redis-cluster-3.redis-cluster.default.svc.cluster.local:6379",
	"redis-cluster-4.redis-cluster.default.svc.cluster.local:6379",
	"redis-cluster-5.redis-cluster.default.svc.cluster.local:6379",
}

const (
	WORLD_SERVER_IP_ALL_IN_ONE = "127.0.0.1"
	WORLD_SERVER_IP_CLUSTER    = "127.0.0.1"
	WORLD_SERVER_IP_K8S        = "gos-world-service.default.svc.cluster.local"
	GAME_DOMAIN                = "gos-game-service.default.svc.cluster.local"
)

const (
	TCP_READ_TIMEOUT    = 60 * time.Second
	HEARTBEAT           = 5 * time.Second
	RPC_REQUEST_TIMEOUT = 5 * time.Second
	GAME_CCU_MAX        = 6000
)

const (
	TIMERTASK_CHECK_DURATION  = time.Second // seconds
	TIMERTASK_TASKS_PER_CHECK = 100         // max tasks fetched per check
	TIMERTASK_MAX_RETRY       = 3           // retry 3 times
)

// Keys for retrive redis data
const (
	RK_GAME_APP_IDS = "__GAME_APP_IDS__"
	RK_SCENE_IDS    = "__SCENE_IDS__"
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
	Packet     uint8
	ListenPort string
}

var TCP_SERVER_GAME = &TCP{
	Packet:     4,
	ListenPort: "4000",
}

/*
Rpc Servers
*/
type Rpc struct {
	ListenNet   string // must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	ListenPort  string
	DialOptions []grpc.DialOption
}

var RPC_FOR_CACHE_MGR = &Rpc{
	ListenNet:   "tcp4",
	ListenPort:  "50051",
	DialOptions: []grpc.DialOption{grpc.WithInsecure()},
}

var RPC_FOR_GAME_APP_MGR = &Rpc{
	ListenNet:   "tcp4",
	ListenPort:  "50052",
	DialOptions: []grpc.DialOption{grpc.WithInsecure()},
}

// Dial addresses are dynamic dispatched by GameAppMgr
var RPC_FOR_GAME_APP_STREAM = &Rpc{
	ListenNet:   "tcp4",
	ListenPort:  "50053",
	DialOptions: []grpc.DialOption{grpc.WithInsecure()},
}

// Dial addresses are dynamic dispatched by GameAppMgr
var RPC_FOR_GAME_APP_RPC = &Rpc{
	ListenNet:   "tcp4",
	ListenPort:  "50054",
	DialOptions: []grpc.DialOption{grpc.WithInsecure()},
}

func GetWorldIP() string {
	switch START_TYPE {
	case START_TYPE_ALL_IN_ONE:
		return WORLD_SERVER_IP_ALL_IN_ONE
	case START_TYPE_CLUSTER:
		return WORLD_SERVER_IP_CLUSTER
	case START_TYPE_K8S:
		return WORLD_SERVER_IP_K8S
	}
	return WORLD_SERVER_IP_K8S
}
