package player_rpc

import (
	"api"
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/game_utils"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/packet"
	"goslib/scene_utils"
	"goslib/session_utils"
	"gslib/player"
	"gslib/routes"
	"sync"
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

func RequestPlayer(targetPlayerId string, encode_method string, params interface{}) (interface{}, error) {
	if gen_server.Exists(targetPlayerId) {
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	session, err := session_utils.Find(targetPlayerId)
	if err != nil {
		return nil, err
	}
	if session.GameAppId == player.CurrentGameAppId {
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	writer := api.Encode(encode_method, params)
	data := writer.GetSendData()
	return crossRequestPlayer(session, data)
}

func RequestPlayerRaw(targetPlayerId string, data []byte) (interface{}, error) {
	if gen_server.Exists(targetPlayerId) {
		encode_method, params := parseData(data)
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	session, err := session_utils.Find(targetPlayerId)
	if err != nil {
		return nil, err
	}
	if session.GameAppId == player.CurrentGameAppId {
		encode_method, params := parseData(data)
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	return crossRequestPlayer(session, data)
}

func internalRequestPlayer(targetPlayerId string, encode_method string, params interface{}) (interface{}, error) {
	handler, err := routes.Route(encode_method)
	if err != nil {
		logger.ERR("internalRequestPlayer failed: ", encode_method, " err: ", err)
		return nil, err
	}
	result, err := player.CallPlayer(targetPlayerId, "handleRPCCall", handler, params)
	if err != nil {
		return nil, err
	}
	return result.(*player.RPCReply).Response, nil
}

func crossRequestPlayer(session *session_utils.Session, data []byte) (interface{}, error) {
	client, err := getClient(session.SceneId)
	if err != nil {
		delClient(session.SceneId)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	reply, err := client.RequestPlayer(ctx, &pb.RequestPlayerRequest{session.AccountId, data})
	if err != nil {
		logger.ERR("RequestPlayer failed: ", err)
		delClient(session.SceneId)
		return nil, err
	}

	_, params := parseData(reply.Data)
	return params, nil
}

func getClient(sceneId string) (pb.RouteConnectGameClient, error) {
	if client, ok := rpcClients.Load(sceneId); ok {
		return client.(pb.RouteConnectGameClient), nil
	}
	client, err := gen_server.Call(PLAYER_RPC_SERVER, "connectScene", sceneId)
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

func parseData(requestData []byte) (decode_method string, params interface{}) {
	reader := packet.Reader(requestData)
	reader.ReadUint16() // read data length
	protocol := reader.ReadUint16()
	decode_method = api.IdToName[protocol]
	params = api.Decode(decode_method, reader)
	return decode_method, params
}
