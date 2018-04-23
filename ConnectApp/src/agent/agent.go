package agent

import (
	"google.golang.org/grpc"
	"connection"
	pb "connectAppProto"
	"time"
	"io"
	"context"
	"goslib/logger"
	"log"
	"goslib/gen_server"
	"sync"
	"strings"
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
func DispatchToGameApp(session *connection.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := GameMgrRpcClient.DispatchGame(ctx, &pb.DispatchGameRequest{
		AccountId: session.AccountId,
		ServerId: session.ServerId,
		SceneId: session.SceneId,
		ConnectId: session.ConnectId,
	})
	if err != nil {
		logger.ERR("could not greet: ", err)
		return err
	}

	logger.DEBUG("Greeting: %s:%s", reply.GetGameAppHost(), reply.GetGameAppPort())

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
		self.doConnectGameAppMgr(gameAppId, addr)
	}
	return nil
}

func (self *Agent) Terminate(reason string) (err error) {
	return nil
}

/*
 * connect to ConnectAppMgr
 */
func (self *Agent)doConnectGameAppMgr(gameAppId string, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
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

	startReceiving(stream)
}

func startReceiving(stream pb.RouteConnectGame_AgentStreamClient) {
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s at point(%d, %d)", in.GetAccountId(), in.GetData())
			proxyToClient(in.AccountId, in.Data)
		}
	}()
}

func proxyToClient(accountId string, data []byte) {
	session := connection.GetSession(accountId)
	session.Connection.SendRawData(data)
}
