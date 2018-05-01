package player

import (
	"goslib/gen_server"
	"api"
	"sync"
	pb "gos_rpc_proto"
	"goslib/logger"
	"goslib/session_utils"
	"goslib/game_utils"
	"gosconf"
	"google.golang.org/grpc"
	"fmt"
	"context"
	"gslib/routes"
	"goslib/scene_utils"
)

/*
   GenServer Callbacks
*/
type PlayerRPC struct {
}

const PLAYER_RPC_SERVER = "__PLAYER_RPC__"
var rpcClients = &sync.Map{}

func StartPlayerRPC() {
	gen_server.Start(PLAYER_RPC_SERVER, new(PlayerRPC))
}

func internalRequestPlayer(targetPlayerId string, encode_method string, params interface{}) (interface{}, error) {
	handler, err := routes.Route(encode_method)
	result, err := CallPlayer(targetPlayerId, "handleRPCCall", handler, params)
	if err != nil {
		return nil, err
	}
	return result.(*RPCReply).response, nil
}

func crossRequestPlayer(session *session_utils.Session, encode_method string, msg interface{}) (interface{}, error) {
	client, err := getClient(session.SceneId)
	if err != nil {
		delClient(session.SceneId)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	writer := api.Encode(encode_method, msg)
	data := writer.GetSendData()
	reply, err := client.RequestPlayer(ctx, &pb.RequestPlayerRequest{session.AccountId, data})
	if err != nil {
		logger.ERR("RequestPlayer failed: ", err)
		delClient(session.SceneId)
		return nil, err
	}

	return parseResponseData(reply.Data), nil
}

func getClient(sceneId string) (pb.RouteConnectGameClient, error) {
	if client, ok := rpcClients.Load(sceneId); ok {
		return client.(pb.RouteConnectGameClient), nil
	}
	client, err := gen_server.Call(SERVER, "connectScene", sceneId)
	if err != nil {
		logger.ERR("connectScene failed: ", err)
		return nil, err
	}
	return client.(pb.RouteConnectGameClient), nil
}

func delClient(gameAppId string) {
	rpcClients.Delete(gameAppId)
}

func (self *PlayerRPC) Init(args []interface{}) (err error) {
	return nil
}

func (self *PlayerRPC) HandleCast(args []interface{}) {
}

func (self *PlayerRPC) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "connectScene" {
		sceneId := args[1].(string)
		if client, ok := rpcClients.Load(sceneId); ok {
			return client, nil
		}
		scene, err := scene_utils.FindScene(sceneId)
		if err != nil {
			return nil, err
		}
		game, err := game_utils.Find(scene.GameAppId)
		if err != nil {
			return nil, err
		}
		client, err := connectGame(sceneId, game)
		return client, err
	}
	return nil, nil
}

func (self *PlayerRPC) Terminate(reason string) (err error) {
	return nil
}

func connectGame(sceneId string, game *game_utils.Game) (pb.RouteConnectGameClient, error) {
	conf := gosconf.RPC_FOR_CONNECT_APP_MGR
	addr := fmt.Sprintf("%s:%s", game.Host, game.Port)
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		logger.ERR("connect AgentMgr failed: ", err)
		return nil, err
	}
	client := pb.NewRouteConnectGameClient(conn)
	rpcClients.Store(sceneId, client)
	return client, nil
}
