package memstore

import (
	"goslib/gen_server"
	"goslib/logger"
	"database/sql"
	"github.com/go-gorp/gorp"
	"time"
	"sync"
)

const PERSISTER_SERVER = "__PERSISTER_SERVER__"
var queueSummary = &sync.Map{}

type SchemaPersistance struct {
	Uuid    string `db:"uuid"`
	Version int    `db:"version"`
}

type PersistTask struct {
    version int64
    sql string
}

type TaskQueue []*PersistTask
type Queue map[string]TaskQueue

/*
   GenServer Callbacks
*/
type DBPersister struct {
	queue Queue
	persistTicker *time.Ticker
}

func StartDBPersister() {
	gen_server.Start(PERSISTER_SERVER, new(DBPersister))
}

func SyncPersistAll() bool {
	gen_server.Call(PERSISTER_SERVER, "SyncPersistAll")
	count := 0
	queueSummary.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count == 0
}

func AddPersistTask(playerId string, version int64, sql string) error {
	logger.INFO("AddPersistTask: ", sql)
	return gen_server.Cast(PERSISTER_SERVER, "AddPersistTask", playerId, version, sql)
}

func IsPersistFinished(playerId string) bool {
	if count, ok := queueSummary.Load(playerId); ok {
		return count.(int) == 0
	}
	return true
}

func EnsurePersisted(playerId string) bool {
	for i := 0; i < 10; i++ {
		if ok := IsPersistFinished(playerId); ok {
			return true
		}
		time.Sleep(time.Second)
	}
	return IsPersistFinished(playerId)
}

func (self *DBPersister) Init(args []interface{}) (err error) {
	self.queue = make(Queue)
	self.persistTicker = time.NewTicker(time.Second)
	go func() {
		for range self.persistTicker.C {
			gen_server.Call(PERSISTER_SERVER, "tickerPersist")
		}
	}()
	return nil
}

func (self *DBPersister) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "AddPersistTask" {
		playerId := args[1].(string)
		version := args[2].(int64)
		sql := args[3].(string)
		self.addTask(playerId, version, sql)
	} else if handle == "tickerPersist" {
		self.tickerPersist()
	}
}

func (self *DBPersister) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "SyncPersistAll" {
		self.tickerPersist()
	}
	return nil, nil
}

func (self *DBPersister) Terminate(reason string) (err error) {
	self.persistTicker.Stop()
	return nil
}

func (self *DBPersister) addTask(playerId string, version int64, sql string) {
	if count, loaded := queueSummary.LoadOrStore(playerId, 1); loaded {
		queueSummary.Store(playerId, count.(int) + 1)
	}
	task := &PersistTask{
		version: version,
		sql: sql,
	}
	var taskQueue TaskQueue
	var ok bool
	if taskQueue, ok = self.queue[playerId]; !ok {
		taskQueue := make(TaskQueue, 0)
		taskQueue = append(taskQueue, task)
		self.queue[playerId] = taskQueue
	}
	taskQueue = append(taskQueue, task)
}

func (self *DBPersister) tickerPersist() {
	dbIns := GetSharedDBInstance()
	finishedPlayers := make([]string, 0)
	for playerId, taskQueue := range self.queue {
		for i := 0; i < len(taskQueue); i++ {
			task := taskQueue[i]
			var schema *SchemaPersistance
			err := dbIns.SelectOne(&schema, "SELECT version from schema_persistances where uuid = ?", playerId)
			if err == sql.ErrNoRows {
				_, err := dbIns.Exec("INSERT INTO `schema_persistances` (uuid,version) VALUES (?,?)", playerId, 0)
				if err != nil {
					logger.ERR("Insert schema_persistances failed: ", err)
					break
				}
			} else {
				if err != nil {
					logger.ERR("Fetch schema_persistances version failed: ", err)
					break
				}
				if task.version <= int64(schema.Version) {
					continue
				}
			}
			if err := executePersistSQL(dbIns, playerId, task); err != nil {
				break
			}
			taskQueue = taskQueue[1:]
		}
		taskCount := len(taskQueue)
		if taskCount == 0 {
			finishedPlayers = append(finishedPlayers, playerId)
			queueSummary.Delete(playerId)
		} else {
			queueSummary.Store(playerId, taskCount)
		}
	}
	for _, playerId := range finishedPlayers {
		delete(self.queue, playerId)
	}
}

func executePersistSQL(dbIns *gorp.DbMap, playerId string, task *PersistTask) error {
	tx, err := dbIns.Db.Begin()
	if err != nil {
		logger.ERR("Start transaction failed: ", err)
		return err
	}
	if _, err := tx.Exec(task.sql); err != nil {
		logger.ERR("ExecutePersist failed: ", err)
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("UPDATE schema_persistances SET version = ? where uuid = ?", task.version, playerId)
	if err != nil {
		logger.ERR("Update schema_persistances version failed: ", err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit();err != nil {
		logger.ERR("ExecutePersistSQL Commit failed: ", err)
		return err
	}
	return nil
}