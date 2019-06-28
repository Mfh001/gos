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
package redis_utils

import (
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/redisdb"
	"strconv"
)

func ToInt32(v string) int32 {
	x, _ := strconv.Atoi(v)
	return int32(x)
}

func ToFloat32(v string) float32 {
	x, _ := strconv.ParseFloat(v, 32)
	return float32(x)
}

func ToInt64(v string) int64 {
	x, _ := strconv.Atoi(v)
	return int64(x)
}

func ToBool(v string) bool {
	return v == "1"
}

func Int32ToString(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

func BatchHGetAll(keys []string) ([]*redis.StringStringMapCmd, error) {
	pipe := redisdb.Instance().Pipeline()
	cmds := make([]*redis.StringStringMapCmd, 0)
	for _, key := range keys {
		cmds = append(cmds, pipe.HGetAll(key))
	}
	if _, err := pipe.Exec(); err != nil {
		return nil, err
	}
	return cmds, nil
}
