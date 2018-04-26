package timertask

import (
	"github.com/ryszard/goskiplist/skiplist"
	"goslib/gen_server"
	"net"
	"fmt"
)

type TimerTask struct {
	Conn net.Conn
}

func Start() {
	addr := "tcp"
	port := 6379
	gen_server.Start(TIMERTASK_SERVER_ID, new(Timertask), addr, port)
}

func Add(key string, delay int, cb func()) {
	gen_server.Cast(TIMERTASK_SERVER_ID, "add", key, delay, cb)
}

func Update(key string, delay int) {
	gen_server.Cast(TIMERTASK_SERVER_ID, "update", key, delay)
}

func Finish(key string) {
	gen_server.Cast(TIMERTASK_SERVER_ID, "finish", key)
}

func Del(key string) {
	gen_server.Cast(TIMERTASK_SERVER_ID, "del", key)
}

func (t *Timertask) Init(args []interface{}) (err error) {
	addr := args[0].(string)
	port := args[1].(string)
	conn, err := redis.Dial(addr, port)
	if err != nil {
		panic(err)
	}
	t.Conn = conn
}

func (t *TimerTask) HandleCall(args []interface{}) interface{} {
	return handleCallAndCast(args)
}

func (t *TimerTask) HandleCast(args []interface{}) {
	handleCallAndCast(args)
}

func (t *TimerTask) handleCallAndCast(args []interface{}) error {
	method := args[0].(string)
	if method == "add" {
		return add(args[1].(string), args[2].(int), args[3].(func()))
	} else if method == "update" {
		return update(args[1].(string), args[2].(int))
	} else if method == "finish" {
		return t.finish(args[1].(string))
	} else if method == "del" {
		return t.del(args[1].(string))
	}
}

func (t *TimerTask) Terminate(args []interface{}) {
}

var KEY string = "TIMERTASK_KEYS"

func mfa_key(key string) string {
	return fmt.Sprintf("%s:mfa_key", key)
}

func (t *TimerTask) add(key string, delay int, cb func()) error {
	return redis_transaction(func() {
		l.Conn.Send("ZADD", KEY, delay, key)
		l.Conn.Send("SET", mfa_key(key), encode(cb))
	})
}

func (t *TimerTask) update(key string, delay int) error {
	count, err := redis.Int(l.Conn.Do("ZREM", KEY, key))
	if count == 1 {
		rep, err := l.Conn.Do("ZADD", KEY, delay, key)
		return err
	}
}

func (t *TimerTask) finish(key string) err {
	count, err := l.Conn.Do("ZREM", KEY, key)
	if err != nil {
		return err
	}
	if count == 1 {
		handle_task(key)
	}
	return nil
}

func (t *TimerTask) del(key string) error {
	_count, err := l.Conn.Do("ZREM", KEY, key)
	return err
}

func redis_transaction(action func()) (interface{}, error) {
	l.Conn.Send("MULTI")
	action()
	return l.Conn.Do("EXEC")
}
