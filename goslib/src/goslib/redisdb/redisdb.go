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
