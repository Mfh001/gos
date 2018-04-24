package redisDB

import (
	"github.com/go-redis/redis"
)

var redisClient *redis.Client

/*
 * Connect Redis Cluster
 */
func Connect(host string, password string, db int) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})
}

func Instance() *redis.Client {
	if redisClient == nil {
		panic("Redis not connected!")
	}
	return redisClient
}
