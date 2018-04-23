package agent

import (
	"goslib/logger"
	pb "connectAppProto"
	"connection"
	"goslib/gen_server"
)

/*
   GenServer Callbacks
*/
type AgentSender struct {
	targetGameApp string
	stream pb.RouteConnectGame_AgentStreamClient
}

func ProxyToGame(session *connection.Session, data []byte) {
	gen_server.Cast(session.ServerId, "sendToGameServer", session.AccountId, data)
}

func (self *AgentSender) Init(args []interface{}) (err error) {
	self.targetGameApp = args[0].(string)
	self.stream = args[1].(pb.RouteConnectGame_AgentStreamClient)
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
			logger.ERR("AgentSender sendMsg error: ", err)
		}
	}
}

func (self *AgentSender) HandleCall(args []interface{}) interface{} {
	return nil
}

func (self *AgentSender) Terminate(reason string) (err error) {
	return nil
}
