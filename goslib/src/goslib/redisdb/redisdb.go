package redisdb

import (
	"github.com/go-redis/redis"
	"gosconf"
	"sync"
)

var serviceClient *redis.Client
var accountClient *redis.Client

var clients = &sync.Map{}

/*
 * Connect Redis Cluster
 */
func Connect(name string, host string, password string, db int) *redis.Client {
	if client, ok := clients.Load(name); ok {
		return client.(*redis.Client)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})
	clients.Store(name, redisClient)
	return redisClient
}

func Instance(name string) *redis.Client {
	if client, ok := clients.Load(name); ok {
		return client.(*redis.Client)
	}
	return nil
}

func InitServiceClient() {
	conf := gosconf.REDIS_FOR_SERVICE
	serviceClient = redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Password,
		DB:       conf.Db,
	})
}

func ServiceInstance() *redis.Client {
	return serviceClient
}

func InitAccountClient() {
	conf := gosconf.REDIS_FOR_ACCOUNT
	accountClient = redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Password,
		DB:       conf.Db,
	})
}

func AccountInstance() *redis.Client {
	return accountClient
}
