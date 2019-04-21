package redisdb

import (
	"flag"
	"github.com/go-redis/redis"
	"gosconf"
)

var clusterClient *redis.ClusterClient

func StartClient() {
	if clusterClient != nil {
		return
	}
	if flag.Lookup("test.v") == nil {
		var addrs []string
		switch gosconf.START_TYPE {
		case gosconf.START_TYPE_K8S:
			addrs = gosconf.REDIS_CLUSTERS_FOR_K8S
			break
		case gosconf.START_TYPE_CLUSTER:
			addrs = gosconf.REDIS_CLUSTERS_FOR_CLUSTER
			break
		case gosconf.START_TYPE_ALL_IN_ONE:
			addrs = gosconf.REDIS_CLUSTERS_FOR_ALL_IN_ONE
			break
		}
		clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: addrs,
		})
	} else {
		clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: gosconf.REDIS_CLUSTERS_FOR_ALL_IN_ONE,
		})
	}
}

func Instance() *redis.ClusterClient {
	return clusterClient
}

// testing code
//var clusterClient *redis.Client
//
//func StartClient() {
//	clusterClient = redis.NewClient(&redis.Options{})
//}
//func Instance() *redis.Client {
//	return clusterClient
//}
