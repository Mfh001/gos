package player_rpc

import (
	"context"
	"fmt"
	"gen/proto"
	"google.golang.org/grpc"
	"gosconf"
	"goslib/api"
	"goslib/game_server/connection"
	"goslib/game_utils"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/packet"
	"goslib/player"
	"goslib/routes"
	"goslib/session_utils"
	"sync"
)

/*
   GenServer Callbacks
*/
type PlayerRPC struct {
}

const PLAYER_RPC_SERVER = "__PLAYER_RPC__"

var rpcClients = &sync.Map{}

func Start() error {
	_, err := gen_server.Start(PLAYER_RPC_SERVER, new(PlayerRPC))
	return err
}

func RpcService(role, playerId, encode_method string, params interface{}) (interface{}, error) {
	return requestPlayer(role, playerId, encode_method, params)
}

func RpcPlayer(playerId, encode_method string, params interface{}) (interface{}, error) {
	return requestPlayer(gosconf.GS_ROLE_DEFAULT, playerId, encode_method, params)
}

func RpcPlayerRaw(playerId string, data []byte) (interface{}, error) {
	return requestPlayerRaw(gosconf.GS_ROLE_DEFAULT, playerId, data)
}

func requestPlayer(role, playerId string, encode_method string, params interface{}) (interface{}, error) {
	if gen_server.Exists(playerId) {
		return internalRequestPlayer(playerId, encode_method, params)
	}
	gameAppId, err := getGameAppId(role, playerId)
	if gameAppId == player.CurrentGameAppId {
		return internalRequestPlayer(playerId, encode_method, params)
	}
	writer, err := api.Encode(encode_method, params)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	data, err := writer.GetSendData(0)
	if err != nil {
		return nil, err
	}
	return crossRequestPlayer(playerId, gameAppId, data)
}

func requestPlayerRaw(role, playerId string, data []byte) (interface{}, error) {
	if gen_server.Exists(playerId) {
		encode_method, params, err := parseData(data)
		if err != nil {
			return nil, err
		}
		return internalRequestPlayer(playerId, encode_method, params)
	}
	gameAppId, err := getGameAppId(role, playerId)
	if err != nil {
		return nil, err
	}
	if gameAppId == player.CurrentGameAppId {
		encode_method, params, err := parseData(data)
		if err != nil {
			return nil, err
		}
		return internalRequestPlayer(playerId, encode_method, params)
	}
	return crossRequestPlayer(playerId, gameAppId, data)
}

func internalRequestPlayer(targetPlayerId string, encode_method string, params interface{}) (interface{}, error) {
	handler, err := routes.Route(encode_method)
	if err != nil {
		logger.ERR("internalRequestPlayer failed: ", encode_method, " err: ", err)
		return nil, err
	}
	result, err := player.CallPlayer(targetPlayerId, &player.RpcCallParams{handler, params})
	if err != nil {
		return nil, err
	}
	return result.(*player.RPCReply).Response, nil
}

func crossRequestPlayer(accountId, gameAppId string, data []byte) (interface{}, error) {
	client, err := getClient(gameAppId)
	if err != nil {
		delClient(gameAppId)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	reply, err := client.RequestPlayer(ctx, &proto.RequestPlayerRequest{
		AccountId: accountId,
		Data:      data,
	})
	if err != nil {
		logger.ERR("RequestPlayer failed: ", err)
		delClient(gameAppId)
		return nil, err
	}

	_, params, err := parseData(reply.Data)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func getClient(gameId string) (proto.GameRpcServerClient, error) {
	if client, ok := rpcClients.Load(gameId); ok {
		return client.(proto.GameRpcServerClient), nil
	}
	client, err := gen_server.Call(PLAYER_RPC_SERVER, &ConnectGameParams{gameId})
	if err != nil {
		logger.ERR("connectGame failed: ", err)
		return nil, err
	}
	return client.(proto.GameRpcServerClient), nil
}

func delClient(gameAppId string) {
	rpcClients.Delete(gameAppId)
}

func (self *PlayerRPC) Init(args []interface{}) (err error) {
	return nil
}

func (self *PlayerRPC) HandleCast(req *gen_server.Request) {
}

func (self *PlayerRPC) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *ConnectGameParams:
		return self.handleConnectGame(params)
	}
	return nil, nil
}

func (self *PlayerRPC) Terminate(reason string) (err error) {
	return nil
}

type ConnectGameParams struct{ gameId string }

func (self *PlayerRPC) handleConnectGame(params *ConnectGameParams) (interface{}, error) {
	game, err := game_utils.Find(params.gameId)
	if err != nil {
		return nil, err
	}
	client, err := connectGame(game)
	return client, err
}

func connectGame(game *game_utils.Game) (proto.GameRpcServerClient, error) {
	conf := gosconf.RPC_FOR_GAME_APP_RPC
	addr := fmt.Sprintf("%s:%s", game.Host, game.RpcPort)
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		logger.ERR("connect AgentMgr failed: ", err)
		return nil, err
	}
	client := proto.NewGameRpcServerClient(conn)
	rpcClients.Store(game.Uuid, client)
	return client, nil
}

func parseData(requestData []byte) (decode_method string, params interface{}, err error) {
	reader := packet.Reader(requestData)
	_, err = reader.ReadDataLength()
	if err != nil {
		return
	}
	_, decode_method, params, err = api.ParseRequestData(reader.RemainData())
	if err != nil {
		logger.ERR("player_rpc parseData failed: ", err)
		return
	}
	return
}

func getGameAppId(role, accountId string) (string, error) {
	session, err := session_utils.Find(accountId)
	if err != nil {
		return "", err
	}
	if session == nil || session.GameAppId == "" {
		game, err := connection.ChooseGameServer(&session_utils.Session{
			GameRole:  role,
			AccountId: accountId,
		})
		if err != nil {
			return "", err
		}
		return game.Uuid, err
	}
	return session.GameAppId, err
}
