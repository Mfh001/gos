package cache_mgr

import (
	"database/sql"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/mysqldb"
	"goslib/redisdb"
	"time"
)

type Task struct {
	Content    string
	Version    int64
	NeedExpire bool
}

type Persister struct {
	queue         map[string]*Task
	persistTicker *time.Ticker
}

const SERVER = "__PERSISTER__"

func StartPersister() {
	gen_server.Start(SERVER, new(Persister))
}

func EnsurePersistered() {
	for {
		count, err := gen_server.Call(SERVER, "remainTasks")
		if err == nil && count.(int) == 0 {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func persistToMySQL(playerId, content string, version int64, needExpire bool) {
	gen_server.Cast(SERVER, "persist", playerId, &Task{
		Content:    content,
		Version:    version,
		NeedExpire: needExpire,
	})
}

func (self *Persister) Init(args []interface{}) (err error) {
	self.queue = make(map[string]*Task)
	self.persistTicker = time.NewTicker(time.Second)
	go func() {
		var err error
		for range self.persistTicker.C {
			_, err = gen_server.Call(SERVER, "tickerPersist")
			if err != nil {
				logger.ERR("call tickerPersist failed: ", err)
			}
		}
	}()
	return nil
}

func (self *Persister) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "persist" {
		playerId := args[1].(string)
		task := args[2].(*Task)
		self.queue[playerId] = task
	}
}

func (self *Persister) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "SyncPersistAll" {
		self.tickerPersist()
	} else if handle == "tickerPersist" {
		self.tickerPersist()
	} else if handle == "remainTasks" {
		return len(self.queue), nil
	}
	return nil, nil
}

func (self *Persister) Terminate(reason string) (err error) {
	return nil
}

func (self *Persister) tickerPersist() {
	for playerId, task := range self.queue {
		if err := persist(playerId, task); err == nil {
			delete(self.queue, playerId)
			if task.NeedExpire {
				_, err = redisdb.Instance().Expire(cacheKey(playerId), CACHE_EXPIRE).Result()
				if err != nil {
					logger.ERR("Persister setexpire failed: ", playerId, err)
				}
			}
		}
	}
}

func persist(playerId string, task *Task) (err error) {
	var lastVersion int
	err = mysqldb.Instance().Db.QueryRow("SELECT updated_at FROM player_datas WHERE uuid=?", playerId).Scan(&lastVersion)
	if err == sql.ErrNoRows {
		query := "INSERT INTO player_datas (uuid, content, updated_at) VALUES (?, ?, ?)"
		_, err = mysqldb.Instance().Db.Exec(query, playerId, task.Content, task.Version)
		return
	}
	if err != nil {
		logger.ERR("Persist data failed: ", playerId, err)
		return
	}
	if int64(lastVersion) < task.Version {
		query := "UPDATE player_datas SET content=?, updated_at=? WHERE uuid=? and updated_at < ?"
		_, err = mysqldb.Instance().Db.Exec(query, task.Content, task.Version, playerId, task.Version)
		return
	}
	return
}
