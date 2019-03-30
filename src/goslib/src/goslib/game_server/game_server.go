package game_server

import (
	"gen/register"
	"gosconf"
	"goslib/broadcast"
	"goslib/game_server/agent"
	"goslib/game_utils"
	"goslib/logger"
	"goslib/mysqldb"
	"goslib/player"
	"goslib/player_data"
	"goslib/player_rpc"
	"goslib/redisdb"
	"goslib/scene_mgr"
	"goslib/timertask"
	"goslib/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strconv"
	"time"
)

type NodeInfo struct {
	Hostname   string
	NodeHost   string
	NodePort   string
	RpcHost    string
	RpcPort    string
	StreamPort string
}

func Start(customRegister func()) {
	go utils.SysRoutine()

	customRegister()

	err := mysqldb.StartClient()
	if err != nil {
		panic(err.Error())
	}
	register.RegisterTables(mysqldb.Instance())

	broadcast.StartMgr()

	scene_mgr.Start()

	player.StartManager()
	player_rpc.Start()

	timertask.Start()

	player_data.Start()

	StartRpcStream()
	agent.Start()

	hostname, err := utils.GetHostname()
	if err != nil {
		logger.ERR("game get host failed: ", err)
		return
	}

	player.CurrentGameAppId = hostname

	for {
		if nodeInfo, err := getNodeInfo(hostname); err == nil {
			if err := reportGameInfo(nodeInfo); err != nil {
				logger.ERR("reportGameInfo failed: ", err)
			}
		}
		time.Sleep(gosconf.HEARTBEAT)
	}
}

func getNodeInfo(hostname string) (*NodeInfo, error) {
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
	logger.INFO("AddGame: ", nodeInfo.Hostname, " Host: ", nodeInfo.NodeHost, " Port: ", nodeInfo.NodePort)

	for {
		heartbeat(app)
		time.Sleep(gosconf.HEARTBEAT)
	}

	return nil
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
	// TODO for k8s health check
	app.Ccu = agent.OnlinePlayers
	app.ActiveAt = time.Now().Unix()
	err := app.Save()

	if err != nil {
		logger.ERR("game heartbeat failed: ", err)
	} else {
		logger.INFO("GameInfo: ", app.Uuid, " NodeHost: ", app.Host, " NodePort: ", app.Port, " ccu: ", app.Ccu)
	}
}
