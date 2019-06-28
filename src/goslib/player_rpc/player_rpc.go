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
package player_rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/mafei198/gos/goslib/api"
	"github.com/mafei198/gos/goslib/game_utils"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/packet"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/routes"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/mafei198/gos/goslib/utils"
	"google.golang.org/grpc"
	"reflect"
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

func RpcService(playerId string, params interface{}) (interface{}, error) {
	return requestPlayer(gosconf.RPC_CALL_NORMAL, playerId, params)
}

func AsyncRpcService(playerId string, params interface{}) {
	go func() {
		defer utils.RecoverPanic("AsyncRpcService")
		if _, err := requestPlayer(gosconf.RPC_CALL_NORMAL, playerId, params); err != nil {
			logger.ERR("AsyncRpcService failed: ", playerId, params, err)
		}
	}()
}

func RpcPlayer(playerId string, params interface{}) (interface{}, error) {
	return requestPlayer(gosconf.RPC_CALL_NORMAL, playerId, params)
}

func AsyncRpcPlayer(playerId, encode_method string, params interface{}) {
	go func() {
		defer utils.RecoverPanic("AsyncRpcPlayer")
		if _, err := requestPlayer(gosconf.RPC_CALL_NORMAL, playerId, params); err != nil {
			logger.ERR("AsyncRpcPlayer failed: ", playerId, encode_method, params, err)
		}
	}()
}

func RpcPlayerRaw(playerId string, data []byte) (interface{}, error) {
	return requestPlayerRaw(gosconf.RPC_CALL_NORMAL, playerId, data)
}

func AsyncProxyData(playerId string, params interface{}) {
	go func() {
		defer utils.RecoverPanic("player_rpc ProxyData")
		if _, err := requestPlayer(gosconf.RPC_CALL_PROXY_DATA, playerId, params); err != nil {
			logger.ERR("AsyncProxyData failed: ", playerId, err)
		}
	}()
}

func AsyncProxyDataRaw(playerId string, data []byte) {
	go func() {
		defer utils.RecoverPanic("player_rpc ProxyData")
		if _, err := requestPlayerRaw(gosconf.RPC_CALL_PROXY_DATA, playerId, data); err != nil {
			logger.ERR("AsyncProxyDataRaw failed: ", playerId, err)
		}
	}()
}

func requestPlayer(category int32, playerId string, params interface{}) (interface{}, error) {
	encode_method := api.EncodeMethod(params)
	if encode_method == "" {
		errMsg := fmt.Sprintf("no encode handle: %s", reflect.TypeOf(params).String())
		return nil, errors.New(errMsg)
	}
	if gen_server.Exists(playerId) {
		return internalRequestPlayer(category, playerId, encode_method, params)
	}
	gameAppId, err := getGameAppId(playerId)
	if gameAppId == player.CurrentGameAppId {
		return internalRequestPlayer(category, playerId, encode_method, params)
	}
	data, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	return crossRequestPlayer(category, playerId, gameAppId, data)
}

func requestPlayerRaw(category int32, playerId string, data []byte) (interface{}, error) {
	if gen_server.Exists(playerId) {
		return internalRequestPlayerRaw(category, playerId, data)
	}
	gameAppId, err := getGameAppId(playerId)
	if err != nil {
		return nil, err
	}
	if gameAppId == player.CurrentGameAppId {
		return internalRequestPlayerRaw(category, playerId, data)
	}
	return crossRequestPlayer(category, playerId, gameAppId, data)
}

func internalRequestPlayerRaw(category int32, targetPlayerId string, data []byte) (interface{}, error) {
	if category == gosconf.RPC_CALL_PROXY_DATA {
		return nil, player.CastPlayer(targetPlayerId, &player.ProxyData{
			AgentId: targetPlayerId,
			Data:    data,
		})
	}
	encode_method, params, err := parseData(data)
	if err != nil {
		return nil, err
	}
	return internalRequestPlayer(category, targetPlayerId, encode_method, params)
}

func internalRequestPlayer(category int32, targetPlayerId string, encode_method string, params interface{}) (interface{}, error) {
	if category == gosconf.RPC_CALL_PROXY_DATA {
		data, err := encodeParams(params)
		if err != nil {
			return nil, err
		}
		return nil, player.CastPlayer(targetPlayerId, &player.ProxyData{
			AgentId: targetPlayerId,
			Data:    data,
		})
	}
	handler, err := routes.Route(encode_method)
	if err != nil {
		logger.ERR("internalRequestPlayer failed: ", encode_method, " err: ", err)
		return nil, err
	}
	logger.INFO("internalRequestPlayer targetPlayerId: ", targetPlayerId, " encode_method: ", encode_method, " params: ", params)
	result, err := player.CallPlayer(targetPlayerId, &player.RpcCallParams{handler, params})
	if err != nil {
		return nil, err
	}
	response := result.(*player.RPCReply).Response
	logger.INFO("Response targetPlayerId: ", targetPlayerId, " params: ", response)
	return response, nil
}

func crossRequestPlayer(category int32, accountId, gameAppId string, data []byte) (interface{}, error) {
	client, err := getClient(gameAppId)
	if err != nil {
		delClient(gameAppId)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	reply, err := client.RequestPlayer(ctx, &proto.RequestPlayerRequest{
		Category:  category,
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
	addr := fmt.Sprintf("%s:%s", game.RpcHost, game.RpcPort)
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

func getGameAppId(accountId string) (string, error) {
	session, err := session_utils.GetPlayerSession(accountId)
	if err != nil {
		return "", err
	}
	return session.GameAppId, err
}

func encodeParams(params interface{}) ([]byte, error) {
	writer, err := api.Encode(params)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	data, err := writer.GetSendData(0)
	return data, err
}
