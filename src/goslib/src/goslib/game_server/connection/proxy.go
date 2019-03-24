package connection

import (
	"context"
	"errors"
	"gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gosconf"
	"goslib/game_server/interfaces"
	"goslib/game_utils"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/session_utils"
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
func ChooseGameServer(session *session_utils.Session) (*game_utils.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()

	reply, err := GameMgrRpcClient.DispatchGame(ctx, &proto.DispatchGameRequest{
		AccountId: session.AccountId,
		SceneId:   session.SceneId,
	})
	if err != nil {
		logger.ERR("DispatchGame failed: ", err)
		return nil, err
	}

	logger.INFO("ChooseGameServer: ", reply.GetGameAppHost())
	session.GameAppId = reply.GetGameAppId()

	return &game_utils.Game{
		Uuid: reply.GetGameAppId(),
		Host: reply.GetGameAppHost(),
		Port: reply.GetGameAppPort(),
	}, err
}

func ConnectGameServer(gameAppId, accountId, roomAppId, roomId string, agent interfaces.AgentBehavior) (proto.GameStreamAgent_GameStreamClient, error) {
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
		"roomAppId": roomAppId,
		"roomId":    roomId,
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

func (self *ProxyManager) HandleCast(args []interface{}) {
}

func (self *ProxyManager) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "ConnectGameApp" {
		gameAppId := args[1].(string)
		game, err := game_utils.Find(gameAppId)
		if err == nil && game != nil {
			addr := strings.Join([]string{game.RpcHost, game.StreamPort}, ":")
			return self.doConnectGameApp(gameAppId, addr)
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
		_, err := gen_server.Call(AGENT_SERVER, "ConnectGameApp", gameAppId)
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
