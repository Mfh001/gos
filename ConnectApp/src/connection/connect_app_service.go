package connection

import (
	"google.golang.org/grpc"
	pb "gosRpcProto"
	"time"
	"context"
	"goslib/logger"
	"goslib/gen_server"
	"sync"
	"strings"
	"goslib/sessionMgr"
	"gosconf"
	"google.golang.org/grpc/metadata"
	"net"
	"io"
)

var GameMgrRpcClient pb.GameDispatcherClient

/*
TODO steam断掉之后重新建立
 */

/*
   GenServer Callbacks
*/
type Agent struct {
}

const AGENT_SERVER = "AGENT_SERVER"

var gameConnMap *sync.Map
var accountStreamMap *sync.Map

func Setup() {
	gameConnMap = new(sync.Map)
	accountStreamMap = new(sync.Map)
	gen_server.Start(AGENT_SERVER, new(Agent))
}

// Request GameAppMgr to dispatch GameApp for session
func DispatchToGameApp(session *sessionMgr.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := GameMgrRpcClient.DispatchGame(ctx, &pb.DispatchGameRequest{
		AccountId: session.AccountId,
		ServerId: session.ServerId,
		SceneId: session.SceneId,
	})
	if err != nil {
		logger.ERR("DispatchGame failed: ", err)
		return err
	}

	session.GameAppId = reply.GetGameAppId()
	session.SceneId = reply.GetSceneId()
	session.Save()

	err = MakeSureConnectedToGame(reply.GetGameAppId(), reply.GetGameAppHost(), reply.GetGameAppPort())

	return err
}

func StartConnToGameStream(gameAppId string, accountId string, rawConn net.Conn) (pb.RouteConnectGame_AgentStreamClient, error) {
	conn := GetGameAppServiceConn(gameAppId)
	client := pb.NewRouteConnectGameClient(conn)
	header := metadata.New(map[string]string{"session": accountId})
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
			logger.INFO("AgentStream received: ", in.AccountId)
			rawConn.Write(in.GetData())
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

func (self *Agent) Init(args []interface{}) (err error) {
	return nil
}

func (self *Agent) HandleCast(args []interface{}) {
}

func (self *Agent) HandleCall(args []interface{}) interface{} {
	handle := args[0].(string)
	if handle == "ConnectGameApp" {
		gameAppId := args[1].(string)
		addr := args[2].(string)
		self.doConnectGameApp(gameAppId, addr)
	}
	return nil
}

func (self *Agent) Terminate(reason string) (err error) {
	return nil
}

/*
 * connect to GameApp Stream
 */
func (self *Agent)doConnectGameApp(gameAppId string, addr string) {
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	logger.INFO("ConnectGameApp: ", addr)
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		logger.ERR("did not connect: ", err)
		return
	}
	gameConnMap.Store(gameAppId, conn)
}

func GetGameAppServiceConn(gameAppId string) *grpc.ClientConn {
	conn, ok := gameConnMap.Load(gameAppId)
	if !ok {
		return nil
	}
	return conn.(*grpc.ClientConn)
}

