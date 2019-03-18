/*
Check timertask every seconds
*/

package timertask

import (
	"api"
	"fmt"
	"github.com/go-redis/redis"
	"gosconf"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/redisdb"
	"gslib/player_rpc"
	"strconv"
	"strings"
	"time"
)

type TimerTask struct {
	taskTicker *time.Ticker
	retry      map[string]int
}

const SERVER = "__TIMERTASK_SERVER__"
const KEY = "TIMERTASK_KEYS"

func Start() {
	gen_server.Start(SERVER, new(TimerTask))
}

func Add(key string, runAt int64, playerId string, encode_method string, params interface{}) error {
	writer, err := api.Encode(encode_method, params)
	if err != nil {
		return err
	}
	data := writer.GetSendData()
	content := fmt.Sprintf("%s:%s", playerId, string(data))
	gen_server.Cast(SERVER, "add", key, runAt, content)
	return nil
}

func Update(key string, runAt int64) {
	gen_server.Cast(SERVER, "update", key, runAt)
}

func Finish(key string) {
	gen_server.Cast(SERVER, "finish", key)
}

func Del(key string) {
	gen_server.Cast(SERVER, "del", key)
}

func (t *TimerTask) Init(args []interface{}) (err error) {
	t.taskTicker = time.NewTicker(gosconf.TIMERTASK_CHECK_DURATION)
	t.retry = make(map[string]int)
	go func() {
		for range t.taskTicker.C {
			gen_server.Call(SERVER, "tickerTask")
		}
	}()
	return nil
}

func (t *TimerTask) HandleCall(args []interface{}) (interface{}, error) {
	err := t.handleCallAndCast(args)
	return nil, err
}

func (t *TimerTask) HandleCast(args []interface{}) {
	t.handleCallAndCast(args)
}

func (t *TimerTask) handleCallAndCast(args []interface{}) error {
	method := args[0].(string)
	if method == "add" {
		key := args[1].(string)
		runAt := args[2].(int64)
		content := args[3].(string)
		return t.add(key, runAt, content)
	} else if method == "update" {
		key := args[1].(string)
		runAt := args[2].(int64)
		return t.update(key, runAt)
	} else if method == "finish" {
		key := args[1].(string)
		return t.finish(key)
	} else if method == "del" {
		key := args[1].(string)
		return t.del(key)
	} else if method == "tickerTask" {
		t.tickerTask()
	}
	return nil
}

func (t *TimerTask) Terminate(reason string) (err error) {
	t.taskTicker.Stop()
	return nil
}

func mfa_key(key string) string {
	return fmt.Sprintf("timertask:%s", key)
}

func (t *TimerTask) add(key string, runAt int64, content string) error {
	if _, err := redisdb.Instance().Set(mfa_key(key), content, 0).Result(); err != nil {
		return err
	}
	member := redis.Z{
		Score:  float64(runAt),
		Member: key,
	}
	if _, err := redisdb.Instance().ZAdd(KEY, member).Result(); err != nil {
		return err
	}
	return nil
}

func (t *TimerTask) update(key string, runAt int64) error {
	score, err := redisdb.Instance().ZScore(KEY, key).Result()
	if err != nil {
		return err
	}
	if score > 0 {
		member := redis.Z{
			Score:  float64(runAt),
			Member: key,
		}
		_, err := redisdb.Instance().ZAdd(KEY, member).Result()
		return err
	}
	return nil
}

func (t *TimerTask) finish(key string) error {
	if err := t.handleTask(key); err != nil {
		count, ok := t.retry[key]
		if ok {
			count = count + 1
			t.retry[key] = count
		} else {
			t.retry[key] = 1
		}
		if count >= gosconf.TIMERTASK_MAX_RETRY {
			delete(t.retry, key)
		} else {
			return err
		}
	}
	return t.del(key)
}

func (t *TimerTask) del(key string) error {
	_, err := redisdb.Instance().Del(mfa_key(key)).Result()
	if err != nil {
		return err
	}
	_, err = redisdb.Instance().ZRem(KEY, key).Result()
	return err
}

func (t *TimerTask) handleTask(key string) error {
	content, err := redisdb.Instance().Get(mfa_key(key)).Result()
	if err != nil {
		return err
	}
	chunks := strings.Split(content, ":")
	_, err = player_rpc.RequestPlayerRaw(chunks[0], []byte(chunks[1]))
	return err
}

func (t *TimerTask) tickerTask() {
	opt := redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.Itoa(int(time.Now().Unix())),
		Offset: 0,
		Count:  gosconf.TIMERTASK_TASKS_PER_CHECK,
	}
	members, err := redisdb.Instance().ZRangeByScoreWithScores(KEY, opt).Result()
	if err != nil {
		logger.ERR("tickerTask failed: ", err)
		return
	}
	for _, member := range members {
		t.finish(member.Member.(string))
	}
}
