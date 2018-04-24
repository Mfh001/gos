package connection

import (
	"goslib/logger"
	pb "gosRpcProto"
	"goslib/gen_server"
	"goslib/sessionMgr"
)

/*
   GenServer Callbacks
*/
type AgentSender struct {
	gameAppId string
	stream pb.RouteConnectGame_AgentStreamClient
}

func StartAgentSender(gameAppId string, stream pb.RouteConnectGame_AgentStreamClient) {
	gen_server.Start(gameAppId, new(AgentSender), gameAppId, stream)
}

func ProxyToGame(session *sessionMgr.Session, data []byte) {
	gen_server.Cast(session.GameAppId, "sendToGameServer", session.AccountId, data)
}

func (self *AgentSender) Init(args []interface{}) (err error) {
	name := args[0].(string)
	logger.INFO("AgentSender started: ", name)
	self.gameAppId = args[1].(string)
	self.stream = args[2].(pb.RouteConnectGame_AgentStreamClient)
	return nil
}

func (self *AgentSender) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "sendToGameServer" {
		accountId := args[1].(string)
		data := args[2].([]byte)
		err := self.stream.Send(&pb.RouteMsg{
			AccountId: accountId,
			Data: data,
		})
		if err != nil {
			logger.ERR("AgentServer sendMsg failed: ", accountId, " err: ", err)
		}
	}
}

func (self *AgentSender) HandleCall(args []interface{}) interface{} {
	return nil
}

func (self *AgentSender) Terminate(reason string) (err error) {
	return nil
}
