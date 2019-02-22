package connection

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/game_utils"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/session_utils"
	"io"
	"net"
	"strings"
	"sync"
)

var GameMgrRpcClient pb.GameDispatcherClient

/*
   GenServer Callbacks
*/
type ProxyManager struct {
}

const AGENT_SERVER = "AGENT_SERVER"

var gameConnMap *sync.Map
var accountStreamMap *sync.Map

func StartProxyManager() {
	gameConnMap = new(sync.Map)
	accountStreamMap = new(sync.Map)
	gen_server.Start(AGENT_SERVER, new(ProxyManager))
}

// Request GameAppMgr to dispatch GameApp for session
func ChooseGameServer(session *session_utils.Session) (*game_utils.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()

	reply, err := GameMgrRpcClient.DispatchGame(ctx, &pb.DispatchGameRequest{
		AccountId: session.AccountId,
		ServerId:  session.ServerId,
	})
	if err != nil {
		logger.ERR("DispatchGame failed: ", err)
		return "", err
	}

	logger.INFO("ChooseGameServer: ", reply.GetGameAppHost())
	session.GameAppId = reply.GetGameAppId()
	session.Save()

	err = MakeSureConnectedToGame(reply.GetGameAppId(), reply.GetGameAppHost(), reply.GetGameAppPort())

	return &game_utils.Game{
		Uuid: reply.GetGameAppId(),
		Host: reply.GetGameAppHost(),
		Port: reply.GetGameAppPort(),
	}, err
}

func ConnectGameServer(gameAppId string, accountId string, agent AgentBehavior) (pb.RouteConnectGame_AgentStreamClient, error) {
	conn := GetGameServerConn(gameAppId)
	client := pb.NewRouteConnectGameClient(conn)
	header := metadata.New(map[string]string{"accountId": accountId})
	ctx := metadata.NewOutgoingContext(context.Background(), header)
	stream, err := client.AgentStream(ctx)
	if err != nil {
		logger.ERR("startPlayerProxyStream failed: ", client, " err:", err)
		return nil, err
	}

	accountStreamMap.Store(accountId, stream)

	// start stream receiver
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
			agent.SendMessage(in.GetData())
		}
		accountStreamMap.Delete(accountId)
		stream.CloseSend()
	}()

	return stream, nil
}

func MakeSureConnectedToGame(gameAppId string, host string, port string) error {
	_, ok := gameConnMap.Load(gameAppId)
	if ok {
		return nil
	}

	addr := strings.Join([]string{host, port}, ":")
	_, err := gen_server.Call(AGENT_SERVER, "ConnectGameApp", gameAppId, addr)
	if err != nil {
		logger.ERR("StartStream failed: ", err)
	}
	return err
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
		addr := args[2].(string)
		self.doConnectGameApp(gameAppId, addr)
	}
	return nil, nil
}

func (self *ProxyManager) Terminate(reason string) (err error) {
	return nil
}

/*
 * connect to GameApp Stream
 */
func (self *ProxyManager) doConnectGameApp(gameAppId string, addr string) {
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	logger.INFO("ConnectGameApp: ", addr)
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		logger.ERR("did not connect: ", err)
		return
	}
	gameConnMap.Store(gameAppId, conn)
}

func GetGameServerConn(gameAppId string) *grpc.ClientConn {
	conn, ok := gameConnMap.Load(gameAppId)
	if !ok {
		return nil
	}
	return conn.(*grpc.ClientConn)
}
