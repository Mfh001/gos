/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package player_data

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	proto_rpc "github.com/mafei198/gos/goslib/gen/proto"
	"github.com/golang/protobuf/proto"
	"github.com/mafei198/gos/goslib/gen/db"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/utils"
	"google.golang.org/grpc"
	"io"
	"net"
	"time"
)

var rpcClient proto_rpc.CacheRpcServerClient

func Start() {
	conf := gosconf.RPC_FOR_CACHE_MGR
	for {
		conn, err := grpc.Dial(net.JoinHostPort(gosconf.GetWorldIP(), conf.ListenPort), conf.DialOptions...)
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
func Load(playerId string) (string, *db.PlayerData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	reply, err := rpcClient.Take(ctx, &proto_rpc.TakeRequest{PlayerId: playerId})
	if err != nil {
		return "", nil, err
	}
	if reply.Data == "" {
		return "", &db.PlayerData{
			User: &db.User{Uuid: playerId},
		}, nil
	}
	checkSum := utils.GetMD5Hash(reply.Data)
	data, err := Decompress(reply.Data)
	return checkSum, data, err
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
func Persist(playerId, lastCheckSum string, playerData *db.PlayerData) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	data, err := Compress(playerData)
	if err != nil {
		return lastCheckSum, err
	}
	checkSum := utils.GetMD5Hash(data)
	if lastCheckSum == checkSum {
		logger.INFO("persist cancel, player_data not change!")
		return lastCheckSum, nil
	}
	_, err = rpcClient.Persist(ctx, &proto_rpc.PersistRequest{
		PlayerId: playerId,
		Data:     data,
		Version:  time.Now().Unix(),
	})
	return checkSum, err
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
		logger.INFO("player_data compress failed: ", err)
		return "", err
	}
	err = writer.Close()
	if err != nil {
		logger.INFO("player_data compress close failed: ", err)
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
