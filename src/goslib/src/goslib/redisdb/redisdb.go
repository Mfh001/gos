package redisdb

import (
	"github.com/go-redis/redis"
	"gosconf"
)

var clusterClient *redis.ClusterClient

func Instance() *redis.ClusterClient {
	if clusterClient == nil {
		clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: gosconf.REDIS_CLUSTERS,
		})
	}
	return clusterClient
}

//
//func Instance() *redis.Client {
//	if clusterClient == nil {
//		clusterClient = redis.NewClient(&redis.Options{
//			Addr: "127.0.0.1:6379",
//		})
//	}
//	return clusterClient
//}
