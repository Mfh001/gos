/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package gosconf

import (
	"google.golang.org/grpc"
	"time"
)

const IS_DEBUG = false

/*
玩家与场景连接模式:
direct:
	玩家直接进入场景进行内部交互
	优势: 交互速度快
	劣势: 单场景同时在线数量有限，取决于硬件性能
proxy:
	玩家通过代理与场景交互
	优势: 场景只处理场景相关逻辑，其他业务逻辑得以分离，可以支持更多的玩家
	劣势: 通讯消息多一层转发，速度较direct模式有所降低
*/
const (
	SCENE_CONNECT_MODE_PROXY = iota
	SCENE_CONNECT_MODE_DIRECT
)

const (
	RPC_CALL_NORMAL int32 = iota
	RPC_CALL_PROXY_DATA
)

const (
	SESSION_EXPIRE_DURATION = 1 * time.Hour
)

var SCENE_CONNECT_MODE = SCENE_CONNECT_MODE_PROXY

const (
	GS_ROLE_DEFAULT = "GS"
	GS_ROLE_MAP     = "MAP"
)

const (
	START_TYPE_ALL_IN_ONE = iota
	START_TYPE_CLUSTER
	START_TYPE_K8S
)

const START_TYPE = START_TYPE_CLUSTER

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

var MAIL_EXPIRE = 24 * time.Hour

var RETRY_BACKOFF = 2 * time.Second

var CONFIG_RELOAD_CHANNEL = "__channel_reload_config__"
var CONFIG_GET_KEY = "__gs_configs__"

var MYSQL_DSN_ALL_IN_ONE = "root@/gos_server_development"
var MYSQL_DSN_CLUSTER = "root@/gos_server_development"
var MYSQL_DSN_K8S = "root:UiP8S9NQJx@tcp(single-mysql.default.svc.cluster.local)/gos_server_development"

var MONGO_DB = "gos-production"

var REDIS_FOR_TEST = "127.0.0.1:6379"

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
	GEN_SERVER_CALL_TIMEOUT = 5 * time.Second
)

const (
	TIMERTASK_CHECK_DURATION  = time.Second // seconds
	TIMERTASK_TASKS_PER_CHECK = 100         // max tasks fetched per check
	TIMERTASK_MAX_RETRY       = 3           // retry 3 times
)

// Keys for retrive redis data
const (
	RK_GAME_APP_IDS      = "__GAME_APP_IDS__"
	RK_SCENE_IDS         = "{scene}.__SCENE_IDS__"
	RK_SERVER_IDS        = "__SERVER_IDS__"
	INITED_MAP_SCENE_IDS = "__INITED_MAP_SCENE_IDS__"
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
