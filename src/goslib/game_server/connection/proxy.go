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

package connection

import (
	"context"
	"errors"
	"github.com/mafei198/gos/goslib/game_server/interfaces"
	"github.com/mafei198/gos/goslib/game_utils"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"strings"
	"sync"
)

var GameMgrRpcClient proto.GameDispatcherClient

/*
   GenServer Callbacks
*/
type ProxyManager struct {
}

const AGENT_SERVER = "AGENT_SERVER"

var gameConnMap *sync.Map

func StartProxyManager() {
	gameConnMap = new(sync.Map)
	gen_server.Start(AGENT_SERVER, new(ProxyManager))
}

// Request GameAppMgr to dispatch GameApp for session
func ChooseGameServer(accountId, sceneId string) (*game_utils.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()

	reply, err := GameMgrRpcClient.DispatchGame(ctx, &proto.DispatchGameRequest{
		AccountId: accountId,
		SceneId:   sceneId,
	})
	if err != nil {
		logger.ERR("DispatchGame failed: ", err)
		return nil, err
	}

	logger.INFO("ChooseGameServer: ", reply.GetGameAppHost())

	return &game_utils.Game{
		Uuid: reply.GetGameAppId(),
		Host: reply.GetGameAppHost(),
		Port: reply.GetGameAppPort(),
	}, err
}

func ConnectGameServer(gameAppId, accountId, sceneId string, agent interfaces.AgentBehavior) (proto.GameStreamAgent_GameStreamClient, error) {
	conn, err := GetGameServerConn(gameAppId)
	if err != nil {
		logger.ERR("GetGameServerConn failed: ", err)
		return nil, err
	}
	if conn == nil {
		return nil, errors.New("game server not exists")
	}
	client := proto.NewGameStreamAgentClient(conn)
	header := metadata.New(map[string]string{
		"accountId": accountId,
		"sceneId":   sceneId,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), header)
	stream, err := client.GameStream(ctx)
	if err != nil {
		logger.ERR("startPlayerProxyStream failed: ", client, " err:", err)
		return nil, err
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				logger.ERR("AgentStream read done: ", err)
				break
			}
			if err != nil {
				logger.ERR("AgentStream failed to receive : ", err)
				break
			}
			logger.INFO("AgentStream received: ", accountId)
			err = agent.SendMessage(in.GetData())
			if err != nil {
				logger.ERR("AgentStream failed to forward msg to client : ", err)
				break
			}
		}
		err := stream.CloseSend()
		if err != nil {
			logger.ERR("proxy close stream failed: ", err)
		}
	}()

	return stream, nil
}

func (self *ProxyManager) Init(args []interface{}) (err error) {
	return nil
}

func (self *ProxyManager) HandleCast(req *gen_server.Request) {
}

type ConnectGameAppParams struct {
	gameAppId string
}

func (self *ProxyManager) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *ConnectGameAppParams:
		game, err := game_utils.Find(params.gameAppId)
		if err == nil && game != nil {
			addr := strings.Join([]string{game.RpcHost, game.StreamPort}, ":")
			return self.doConnectGameApp(params.gameAppId, addr)
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (self *ProxyManager) Terminate(reason string) (err error) {
	return nil
}

/*
 * connect to GameApp Stream
 */
func (self *ProxyManager) doConnectGameApp(gameAppId string, addr string) (*grpc.ClientConn, error) {
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	logger.INFO("ConnectGameApp: ", addr)
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		logger.ERR("did not connect: ", err)
		return nil, err
	}
	gameConnMap.Store(gameAppId, conn)
	return conn, nil
}

func GetGameServerConn(gameAppId string) (*grpc.ClientConn, error) {
	conn, ok := gameConnMap.Load(gameAppId)
	if !ok {
		_, err := gen_server.Call(AGENT_SERVER, &ConnectGameAppParams{gameAppId})
		if err != nil {
			logger.ERR("StartStream failed: ", err)
			return nil, err
		}
		conn, ok = gameConnMap.Load(gameAppId)
		if !ok {
			return nil, nil
		}
	}
	return conn.(*grpc.ClientConn), nil
}
