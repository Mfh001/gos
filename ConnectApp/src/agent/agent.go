package agent

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

var streamMap *sync.Map

func Setup() {
	streamMap = new(sync.Map)
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

func MakeSureConnectedToGame(gameAppId string, host string, port string) error {
	addr := strings.Join([]string{host, port}, ":")
	_, ok := streamMap.Load(gameAppId)
	if ok {
		return nil
	}

	_, err := gen_server.Call(AGENT_SERVER, "ConnectGameAppMgr", gameAppId, addr)
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
	if handle == "ConnectGameAppMgr" {
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
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		logger.ERR("did not connect: ", err)
	}

	client := pb.NewRouteConnectGameClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.AgentStream(ctx)
	if err != nil {
		logger.ERR(client, ".RouteChat(_) = _, ", err)
	}

	streamMap.Store(gameAppId, stream)

	// start stream sender
	StartAgentSender(gameAppId, stream)

	// start stream receiver
	StartReceiving(stream)
}
