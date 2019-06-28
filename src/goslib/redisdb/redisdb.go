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
package redisdb

import "github.com/go-redis/redis"

// cluster mode
/*
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
*/

// standalone mode
var clusterClient *redis.Client

func StartClient() {
	clusterClient = redis.NewClient(&redis.Options{})
}
func Instance() *redis.Client {
	return clusterClient
}
