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

func Start() {
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
	writer, err := api.Encode(encode_method, params)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	data, err := writer.GetSendData(0)
	if err != nil {
		return nil, err
	}
	return crossRequestPlayer(session, data)
}

func RequestPlayerRaw(targetPlayerId string, data []byte) (interface{}, error) {
	if gen_server.Exists(targetPlayerId) {
		encode_method, params, err := parseData(data)
		if err != nil {
			return nil, err
		}
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	session, err := session_utils.Find(targetPlayerId)
	if err != nil {
		return nil, err
	}
	if session.GameAppId == "" {
		if _, err := connection.ChooseGameServer(session); err != nil {
			return nil, err
		}
		session, err = session_utils.Find(targetPlayerId)
		if err != nil {
			return nil, err
		}
	}
	if session.GameAppId == player.CurrentGameAppId {
		encode_method, params, err := parseData(data)
		if err != nil {
			return nil, err
		}
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
	result, err := player.CallPlayer(targetPlayerId, &player.RpcCallParams{handler, params})
	if err != nil {
		return nil, err
	}
	return result.(*player.RPCReply).Response, nil
}

func crossRequestPlayer(session *session_utils.Session, data []byte) (interface{}, error) {
	client, err := getClient(session.GameAppId)
	if err != nil {
		delClient(session.GameAppId)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	reply, err := client.RequestPlayer(ctx, &proto.RequestPlayerRequest{
		AccountId: session.AccountId,
		Data:      data,
	})
	if err != nil {
		logger.ERR("RequestPlayer failed: ", err)
		delClient(session.GameAppId)
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
	client, err := gen_server.Call(PLAYER_RPC_SERVER,&ConnectGameParams{gameId})
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

func (self *PlayerRPC) HandleCast(msg interface{}) {
}

func (self *PlayerRPC) HandleCall(msg interface{}) (interface{}, error) {
	switch params := msg.(type) {
	case *ConnectGameParams:
		return self.handleConnectGame(params)
	}
	return nil, nil
}

func (self *PlayerRPC) Terminate(reason string) (err error) {
	return nil
}

type ConnectGameParams struct { gameId string }
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
