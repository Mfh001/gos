package player

import (
	pb "gosRpcProto"
	"goslib/gen_server"
	"goslib/logger"
)

/*
   GenServer Callbacks
*/
type GameAgentSender struct {
	stream pb.RouteConnectGame_AgentStreamServer
}

func StartGameAgentSender(connectAppId string, stream pb.RouteConnectGame_AgentStreamServer) {
	gen_server.Start(connectAppId, new(GameAgentSender), stream)
}

func ProxyToConnect(connectAppId string, accountId string, data []byte) {
	gen_server.Cast(connectAppId, "ProxyToConnect", accountId, data)
}

func (self *GameAgentSender) Init(args []interface{}) (err error) {
	self.stream = args[1].(pb.RouteConnectGame_AgentStreamServer)
	logger.INFO("GameAgentSender started!")
	return nil
}

func (self *GameAgentSender) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "ProxyToConnect" {
		accountId := args[1].(string)
		data := args[2].([]byte)
		err := self.stream.Send(&pb.RouteMsg{
			accountId,
			data,
		})
		if err != nil {
			logger.ERR("GameAgentSender ProxyToConnect failed: ", accountId, " err: ", err)
		}
	}
}

func (self *GameAgentSender) HandleCall(args []interface{}) interface{} {
	return nil
}

func (self *GameAgentSender) Terminate(reason string) (err error) {
	return nil
}