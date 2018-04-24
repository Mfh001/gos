package connection

import (
	"io"
	pb "gosRpcProto"
	"goslib/logger"
)

func StartReceiving(stream pb.RouteConnectGame_AgentStreamClient) {
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				logger.ERR("AgentStream read done: ", err)
				return
			}
			if err != nil {
				logger.ERR("AgentStream failed to receive : ", err)
			}
			proxyToClient(in.AccountId, in.Data)
		}
	}()
}

func proxyToClient(accountId string, data []byte) {
	SendRawData(accountId, data)
}
