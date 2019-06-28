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
package cache_mgr

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/database"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/mysqldb"
	"github.com/mafei198/gos/goslib/redisdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"net"
	"strings"
	"time"
)

var grpcServer *grpc.Server

type CacheMgr struct {
}

type collectionMap struct {
	PlayerData *mongo.Collection
}

var C collectionMap

const CACHE_EXPIRE = 1 * time.Hour

func Start() {
	DB := database.Instance().Database(gosconf.MONGO_DB)
	C.PlayerData = DB.Collection("player_datas")
	StartPersister()

	conf := gosconf.RPC_FOR_CACHE_MGR
	lis, err := net.Listen(conf.ListenNet, net.JoinHostPort("", conf.ListenPort))
	logger.INFO("CacheRpcServer lis: ", conf.ListenNet, " port: ", conf.ListenPort)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}

	grpcServer = grpc.NewServer()
	proto.RegisterCacheRpcServerServer(grpcServer, &CacheMgr{})

	err = mysqldb.StartClient()
	if err != nil {
		logger.ERR("Start CacheRpcServer failed: ", err)
		panic(err)
	}

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			logger.ERR("Start CacheRpcServer failed: ", err)
			panic(err)
		}
	}()
}

func Stop() {
	grpcServer.GracefulStop()
	EnsurePersistered()
}

func (self *CacheMgr) Take(ctx context.Context, in *proto.TakeRequest) (*proto.TakeReply, error) {
	logger.INFO("cache Take: ", in.PlayerId)
	content, err := getFromRedis(in.PlayerId)
	if err == redis.Nil {
		content, err = getFromMySQL(in.PlayerId)
		if err == sql.ErrNoRows {
			return &proto.TakeReply{}, nil
		}
		if err != nil {
			logger.ERR("Take PlayerData query MySQL failed: ", err)
			return nil, err
		}
		return &proto.TakeReply{Data: content}, nil
	}

	if err != nil {
		logger.ERR("Take PlayerData from redis failed: ", in.PlayerId, err)
		return nil, err
	}

	if err = delFromRedis(in.PlayerId); err != nil {
		logger.ERR("cache_mgr del from redis failed: ", err)
	}

	return &proto.TakeReply{Data: content}, nil
}

func (self *CacheMgr) Return(ctx context.Context, in *proto.ReturnRequest) (*proto.ReturnReply, error) {
	logger.INFO("cache Return: ", in.PlayerId)
	if err := persistToRedis(in.PlayerId, in.Data); err != nil {
		logger.ERR("Return PlayerData failed: ", in.PlayerId, err)
		return &proto.ReturnReply{Success: false}, err
	}

	persistToMySQL(in.PlayerId, in.Data, in.Version, true)

	return &proto.ReturnReply{Success: true}, nil
}

func (self *CacheMgr) Persist(ctx context.Context, in *proto.PersistRequest) (*proto.PersistReply, error) {
	logger.INFO("cache Persist: ", in.PlayerId)
	persistToMySQL(in.PlayerId, in.Data, in.Version, false)
	return &proto.PersistReply{Success: true}, nil
}

func cacheKey(playerId string) string {
	return strings.Join([]string{"player_data", playerId}, ":")
}

func getFromMySQL(playerId string) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	filter := bson.D{
		{"_id", playerId},
	}
	single := C.PlayerData.FindOne(ctx, filter)
	err := single.Err()
	if err != nil {
		return "", single.Err()
	}
	var result bson.M
	if err := single.Decode(&result); err != nil {
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return "", nil
		} else {
			return "", err
		}
	}
	content := result["Content"].(string)
	return content, nil
}

func persistToRedis(playerId, content string) error {
	key := cacheKey(playerId)
	_, err := redisdb.Instance().Set(key, content, 0).Result()
	return err
}

func getFromRedis(playerId string) (string, error) {
	key := cacheKey(playerId)
	return redisdb.Instance().Get(key).Result()
}

func delFromRedis(playerId string) error {
	key := cacheKey(playerId)
	_, err := redisdb.Instance().Del(key).Result()
	return err
}

type PlayerData struct {
	Uuid      string
	Content   string
	UpdatedAt int64
}

func decodeDBRec(result bson.M) *PlayerData {
	return &PlayerData{
		Uuid:      result["_id"].(string),
		Content:   result["Content"].(string),
		UpdatedAt: result["UpdatedAt"].(int64),
	}
}
