package ConnectAppMgr

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "connectAppProto"
	"google.golang.org/grpc/reflection"
	"ConnectAppMgr/connectApp"
	"goslib/redisDB"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type connectAppMgr struct{

}

func Start() {
	connectApp.Start()
	startConnectAppMgrRpc()
}

func startConnectAppMgrRpc() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer()
	pb.RegisterDispatcherServer(rpcServer, &connectAppMgr{})
	// Register reflection service on gRPC server.
	reflection.Register(rpcServer)
	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *connectAppMgr) DispatchPlayer(ctx context.Context, in *pb.DispatchRequest) (*pb.DispatchReply, error) {
	appId, host, port, err := connectApp.Dispatch(in.AccountId, in.GroupId)

	return  &pb.DispatchReply{
		ConnectAppId: appId,
		ConnectAppHost: host,
		ConnectAppPort: port,
	}, err
}
