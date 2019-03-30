package player_data

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"gen/db"
	proto_rpc "gen/proto"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"gosconf"
	"goslib/logger"
	"io"
	"time"
)

var rpcClient proto_rpc.CacheRpcServerClient

func Start() {
	conf := gosconf.RPC_FOR_CACHE_MGR
	for {
		conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
		if err != nil {
			logger.ERR("Start connect to cache_mgr failed: ", err)
			time.Sleep(gosconf.HEARTBEAT)
			continue
		}
		rpcClient = proto_rpc.NewCacheRpcServerClient(conn)
		return
	}
}

// Load PlayerData from cache_mgr
func Load(playerId string) (*db.PlayerData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	reply, err := rpcClient.Take(ctx, &proto_rpc.TakeRequest{PlayerId: playerId})
	if err != nil {
		return nil, err
	}
	if reply.Data == "" {
		return &db.PlayerData{
			User: &db.User{Uuid: playerId},
		}, nil
	}
	return Decompress(reply.Data)
}

// Return PlayerData to cache_mgr
func Return(playerId string, playerData *db.PlayerData) error {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	data, err := Compress(playerData)
	if err != nil {
		return err
	}
	_, err = rpcClient.Return(ctx, &proto_rpc.ReturnRequest{
		PlayerId: playerId,
		Data:     data,
		Version:  time.Now().Unix(),
	})
	return err
}

// Persist PlayerData to cache_mgr
func Persist(playerId string, playerData *db.PlayerData) error {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	data, err := Compress(playerData)
	if err != nil {
		return err
	}
	_, err = rpcClient.Persist(ctx, &proto_rpc.PersistRequest{
		PlayerId: playerId,
		Data:     data,
		Version:  time.Now().Unix(),
	})
	return err
}

func Compress(playerData *db.PlayerData) (string, error) {
	data, err := proto.Marshal(playerData)
	if err != nil {
		return "", err
	}

	var in bytes.Buffer
	writer := zlib.NewWriter(&in)
	_, err = writer.Write(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(in.Bytes()), nil
}

func Decompress(content string) (*db.PlayerData, error) {
	rawContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewBuffer(rawContent)
	readCloser, err := zlib.NewReader(reader)
	if err != nil {
		return nil, err
	}

	var replyData bytes.Buffer
	_, err = io.Copy(&replyData, readCloser)
	_ = readCloser.Close()
	if err != nil {
		return nil, err
	}

	playerData := &db.PlayerData{}
	err = proto.Unmarshal(replyData.Bytes(), playerData)
	return playerData, err
}
