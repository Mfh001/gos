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
package game_server

import (
	"github.com/mafei198/gos/goslib/broadcast"
	"github.com/mafei198/gos/goslib/config_data"
	"github.com/mafei198/gos/goslib/game_server/agent"
	"github.com/mafei198/gos/goslib/game_utils"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/mem_data"
	"github.com/mafei198/gos/goslib/mysqldb"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/player_data"
	"github.com/mafei198/gos/goslib/player_rpc"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/scene_mgr"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/mafei198/gos/goslib/shared_data"
	"github.com/mafei198/gos/goslib/timertask"
	"github.com/mafei198/gos/goslib/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type NodeInfo struct {
	Role       string
	Hostname   string
	NodeHost   string
	NodePort   string
	RpcHost    string
	RpcPort    string
	StreamPort string
}

var Role string
var customShutdown func()

func Start(role string, customRegister func(), afterStart func(), afterShutdown func()) {
	Role = role
	customShutdown = afterShutdown
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go utils.SysRoutine()

	customRegister()

	session_utils.Start()
	shared_data.StartMgr()

	mem_data.Start()

	config_data.Load()

	err := mysqldb.StartClient()
	if err != nil {
		panic(err)
	}

	broadcast.StartMgr()

	if err = scene_mgr.Start(); err != nil {
		panic(err)
	}

	if err = player.StartManager(); err != nil {
		panic(err)
	}
	if err := player_rpc.Start(); err != nil {
		panic(err)
	}

	player_data.Start()

	StartRpcStream()
	agent.Start()

	hostname, err := utils.GetHostname()
	if err != nil {
		logger.ERR("game get host failed: ", err)
		return
	}

	if err = timertask.Start(hostname); err != nil {
		logger.ERR("start timertask failed: ", err)
		return
	}

	player.CurrentGameAppId = hostname

	if afterStart != nil {
		afterStart()
	}

	go func() {
		for {
			if nodeInfo, err := getNodeInfo(hostname); err == nil {
				if err := reportGameInfo(nodeInfo); err != nil {
					logger.ERR("reportGameInfo failed: ", err)
				}
			}
			time.Sleep(gosconf.HEARTBEAT)
		}
	}()

	<-stopChan // wait for SIGINT or SIGTERM
	logger.INFO("Shutting down game server...")

	shutdown()

	logger.INFO("game server stopped")
}

func shutdown() {
	// Stop tcp acceptor
	logger.INFO("Stop acceptor")
	agent.StopAcceptor()

	// Stop receiving requests
	logger.INFO("Stop accept message")
	agent.StopAcceptMsg()

	// Stop timertask
	logger.INFO("Stop timertask")
	if err := timertask.Stop(); err != nil {
		logger.ERR("timertask stop failed: ", err)
	}

	// Stop players
	logger.INFO("Stopping players")
	player.EnsureShutdown()

	// Stop custom services
	logger.INFO("Stopping custom services")
	customShutdown()

	// Stop shared_data
	logger.INFO("Stopping shared_data")
	shared_data.EnsureShutdown()
}

func getNodeInfo(hostname string) (*NodeInfo, error) {
	switch gosconf.START_TYPE {
	case gosconf.START_TYPE_ALL_IN_ONE:
		return getNodeInfoForAllInOne(hostname)
	case gosconf.START_TYPE_CLUSTER:
		return getNodeInfoForCluster(hostname)
	case gosconf.START_TYPE_K8S:
		return getNodeInfoForK8s(hostname)
	}
	return nil, nil
}

func getNodeInfoForAllInOne(hostname string) (*NodeInfo, error) {
	rpcHost := "127.0.0.1"
	rpcPort := gosconf.RPC_FOR_GAME_APP_RPC.ListenPort
	streamPort := gosconf.RPC_FOR_GAME_APP_STREAM.ListenPort

	return &NodeInfo{
		Role:       Role,
		Hostname:   hostname,
		NodeHost:   "127.0.0.1",
		NodePort:   agent.AgentPort,
		RpcHost:    rpcHost,
		RpcPort:    rpcPort,
		StreamPort: streamPort,
	}, nil
}

func getNodeInfoForCluster(hostname string) (*NodeInfo, error) {
	rpcHost, err := utils.GetLocalIp()
	if err != nil {
		logger.ERR("get localIP failed: ", err)
		panic(err)
	}
	rpcPort := gosconf.RPC_FOR_GAME_APP_RPC.ListenPort
	streamPort := gosconf.RPC_FOR_GAME_APP_STREAM.ListenPort

	nodeHost, err := utils.GetPublicIP()
	if err != nil {
		logger.ERR("get publicIP failed: ", err)
		panic(err)
	}

	return &NodeInfo{
		Role:       Role,
		Hostname:   hostname,
		NodeHost:   nodeHost,
		NodePort:   agent.AgentPort,
		RpcHost:    rpcHost,
		RpcPort:    rpcPort,
		StreamPort: streamPort,
	}, nil
}

func getNodeInfoForK8s(hostname string) (*NodeInfo, error) {
	var externalIP string
	var internalIP string
	var nodePort string
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.ERR("Init k8s rest failed: ", err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.ERR("Init k8s client failed: ", err)
		return nil, err
	}
	service, err := clientset.CoreV1().Services("default").Get(hostname, metav1.GetOptions{})
	if err != nil {
		logger.ERR("get service failed: ", err)
		return nil, err
	}
	pod, err := clientset.CoreV1().Pods("default").Get(hostname, metav1.GetOptions{})
	if err != nil {
		logger.ERR("get pod failed: ", err)
		return nil, err
	}
	node, err := clientset.CoreV1().Nodes().Get(pod.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		logger.ERR("get nod failed: ", err)
		return nil, err
	}

	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			internalIP = address.Address
		}
		if address.Type == "ExternalIP" {
			externalIP = address.Address
		}
	}

	for _, port := range service.Spec.Ports {
		nodePort = strconv.Itoa(int(port.NodePort))
		break
	}

	rpcHost := hostname + "." + gosconf.GAME_DOMAIN
	rpcPort := gosconf.RPC_FOR_GAME_APP_RPC.ListenPort
	streamPort := gosconf.RPC_FOR_GAME_APP_STREAM.ListenPort

	nodeHost := externalIP
	if nodeHost == "" {
		nodeHost = internalIP
	}

	logger.INFO("Hostname: ", hostname, "InternalIP: ", internalIP, "ExternalIP: ", externalIP, " nodeHost: ", nodeHost, " nodePort: ", nodePort)

	return &NodeInfo{
		Role:       Role,
		Hostname:   hostname,
		NodeHost:   nodeHost,
		NodePort:   nodePort,
		RpcHost:    rpcHost,
		RpcPort:    rpcPort,
		StreamPort: streamPort,
	}, nil
}

func reportGameInfo(nodeInfo *NodeInfo) error {
	app, err := addGame(nodeInfo)
	if err != nil {
		logger.ERR("addGame failed: ", err)
		return err
	}
	logger.INFO("AddGame: ", nodeInfo.Hostname, " Host: ", nodeInfo.NodeHost, " Port: ", nodeInfo.NodePort, " RpcHost: ", nodeInfo.RpcHost, " RpcPort: ", nodeInfo.RpcPort)

	for {
		heartbeat(app)
		time.Sleep(gosconf.HEARTBEAT)
	}
}

func addGame(nodeInfo *NodeInfo) (*game_utils.Game, error) {
	app := &game_utils.Game{
		Uuid:       nodeInfo.Hostname,
		Host:       nodeInfo.NodeHost,
		Port:       nodeInfo.NodePort,
		RpcHost:    nodeInfo.RpcHost,
		RpcPort:    nodeInfo.RpcPort,
		StreamPort: nodeInfo.StreamPort,
		ActiveAt:   time.Now().Unix(),
	}
	_, err := redisdb.Instance().SAdd(gosconf.RK_GAME_APP_IDS, app.Uuid).Result()
	if err != nil {
		return nil, err
	}
	err = app.Save()
	return app, err
}

func heartbeat(app *game_utils.Game) {
	app.Ccu = agent.OnlinePlayers
	app.ActiveAt = time.Now().Unix()
	err := app.Save()

	if err != nil {
		logger.ERR("game heartbeat failed: ", err)
	} else {
		logger.INFO("GameInfo: ", app.Uuid, " NodeHost: ", app.Host, " NodePort: ", app.Port, " ccu: ", app.Ccu)
		utils.PrintMemUsage()
	}
}
