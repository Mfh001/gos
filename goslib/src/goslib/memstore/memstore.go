package memstore

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"goslib/logger"
)

type Filter func(elem interface{}) bool

type Store map[string]interface{}

type TableStatus map[string]int8
type StoreStatus map[string]TableStatus

type dataLoader func(modelName string, ets *MemStore)

var dataLoaderMap = map[string]dataLoader{}

type MemStore struct {
	store Store
	storeStatus StoreStatus
	Db    *gorp.DbMap
	Ctx   interface{}
}

/*
 * Memory data status
 */
const (
	STATUS_ORIGIN = 1
	STATUS_CREATE = 2
	STATUS_UPDATE = 3
	STATUS_DELETE = 4
)

var sharedDBInstance *gorp.DbMap

func InitDB() {
	db, err := sql.Open("mysql", "root:@/gos_server_development")
	if err != nil {
		panic(err.Error())
	}
	sharedDBInstance = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
}

func GetSharedDBInstance() *gorp.DbMap {
	return sharedDBInstance
}

func New(ctx interface{}) *MemStore {
	e := &MemStore{
		store: make(Store),
		storeStatus: make(StoreStatus),
		Db:    GetSharedDBInstance(),
		Ctx:   ctx,
	}
	return e
}

func RegisterDataLoader(modelName string, loader dataLoader) {
	dataLoaderMap[modelName] = loader
}

func (e *MemStore) LoadData(modelName, playerId string) {
	handler, ok := dataLoaderMap[modelName]
	if ok {
		handler(playerId, e)
	}
}

func (e *MemStore) Load(namespaces []string, key string, value interface{}) {
	ctx := e.makeCtx(namespaces)
	ctx[key] = value
}

func (e *MemStore) Get(namespaces []string, key string) interface{} {
	if ctx := e.getCtx(namespaces); ctx != nil {
		return ctx[key]
	} else {
		return nil
	}
}

func (e *MemStore) Set(namespaces []string, key string, value interface{}) {
	ctx := e.makeCtx(namespaces)
	if ctx[key] == nil {
		e.UpdateStatus(namespaces[len(namespaces) - 1], key, STATUS_CREATE)
	} else {
		e.UpdateStatus(namespaces[len(namespaces) - 1], key, STATUS_UPDATE)
	}
	ctx[key] = value
}

func (e *MemStore) Del(namespaces []string, key string) {
	if ctx := e.getCtx(namespaces); ctx != nil {
		e.UpdateStatus(namespaces[len(namespaces) - 1], key, STATUS_DELETE)
		delete(ctx, key)
	}
}

func (e *MemStore) Find(namespaces []string, filter Filter) interface{} {
	if ctx := e.getCtx(namespaces); ctx != nil {
		for _, v := range ctx {
			if filter(v) {
				return v
			}
		}
	}
	return nil
}

func (e *MemStore) Select(namespaces []string, filter Filter) interface{} {
	var elems []interface{}
	if ctx := e.getCtx(namespaces); ctx != nil {
		for _, v := range ctx {
			if filter(v) {
				elems = append(elems, v)
			}
		}
		return elems
	}
	return elems
}

func (e *MemStore) Count(namespaces []string) int {
	if ctx := e.getCtx(namespaces); ctx != nil {
		return len(ctx)
	} else {
		return 0
	}
}

/*
 * Persist all tables in []string{"models"} namespaces
 * Example: Persist([]string{"models"})
 */
func (e *MemStore) Persist(namespaces []string) {
	trans, err := e.Db.Begin()
	if err != nil {
		logger.ERR("Persist Begin failed: ", err)
		return
	}
	for tableName, tableCtx := range e.getCtx(namespaces) {
		statusMap, ok := e.tableStatus(tableName)
		if ok {
			executeSql(trans, tableName, statusMap, tableCtx.(Store))
		}
	}
	err = trans.Commit()
	if err != nil {
		logger.ERR("Persist Commit failed: ", err)
		return
	}
	e.cleanStatus()
}

func (e *MemStore) getCtx(namespaces []string) Store {
	var ctx Store = nil
	for _, namespace := range namespaces {
		if ctx == nil {
			vctx, ok := e.store[namespace]
			if !ok {
				return nil
			}
			ctx = vctx.(Store)
		} else {
			vctx, ok := ctx[namespace]
			if !ok {
				return nil
			}
			ctx = vctx.(Store)
		}
	}
	return ctx
}

func (e *MemStore) makeCtx(namespaces []string) Store {
	var ctx Store = nil
	for _, namespace := range namespaces {
		if ctx == nil {
			vctx, ok := e.store[namespace]
			if !ok {
				ctx = make(Store)
				e.store[namespace] = ctx
			} else {
				ctx = vctx.(Store)
			}
		} else {
			vctx, ok := ctx[namespace]
			if !ok {
				vctx = make(Store)
				ctx[namespace] = vctx
			}
			ctx = vctx.(Store)
		}
	}
	return ctx
}

func (e *MemStore) UpdateStatus(table string, key string, status int8) {
	tableStatus, ok := e.storeStatus[table]
	if !ok {
		tableStatus = make(TableStatus)
		e.storeStatus[table] = tableStatus
	}
	tableStatus[key] = status
}

func (e *MemStore) getStatus(table string, key string) int8 {
	tableStatus, ok := e.storeStatus[table]
	if !ok {
		return STATUS_ORIGIN
	}

	status, ok := tableStatus[key]
	if !ok {
		return STATUS_ORIGIN
	}

	return status
}

/*
 * table status map[]int
 */
func (e *MemStore) tableStatus(table string) (TableStatus, bool) {
	tableStatus, ok := e.storeStatus[table]
	return tableStatus, ok
}

func (e *MemStore) cleanStatus() {
	e.storeStatus = make(StoreStatus)
}

func executeSql(trans *gorp.Transaction, tableName string, status TableStatus, tableCtx Store) {
	for k, v := range status {
		switch v {
		case STATUS_UPDATE:
			fmt.Println("STATUS_UPDATE: ", reflect.ValueOf(tableCtx[k]).Elem().FieldByName("Data").Interface())
			_, err := trans.Update(reflect.ValueOf(tableCtx[k]).Elem().FieldByName("Data").Interface())
			if err != nil {
				panic(err.Error())
			}
		case STATUS_DELETE:
			_, err := trans.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `uuid`='%s'", tableName, k))
			if err != nil {
				panic(err.Error())
			}
		case STATUS_CREATE:
			err := trans.Insert(reflect.ValueOf(tableCtx[k]).Elem().FieldByName("Data").Interface())
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

//func updateRec(tableName string, Rec interface{}) {
//	if tableName == "users" {
//		user := Rec.(*models.UserModel).Data
//		return fmt.Sprintf("update ")
//	}
//}
