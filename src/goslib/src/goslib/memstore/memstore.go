package memstore

import (
	"database/sql"
	"errors"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	. "goslib/base_model"
	"goslib/logger"
	"time"
)

type Filter func(elem interface{}) bool

type Store map[string]interface{}

type ModelStatus struct {
	Origin  int8
	Current int8
}
type TableStatus map[string]*ModelStatus
type StoreStatus map[string]TableStatus

type dataLoader func(modelName string, ets *MemStore)

var dataLoaderMap = map[string]dataLoader{}

type MemStore struct {
	playerId    string
	store       Store
	storeStatus StoreStatus
	dataLoaded  map[string]bool
	Db          *gorp.DbMap
	Ctx         interface{}
}

var sharedDBInstance *gorp.DbMap

func StartDB() error {
	db, err := sql.Open("mysql", "root:@/gos_server_development")
	if err != nil {
		return err
	}
	sharedDBInstance = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	return nil
}

func GetSharedDBInstance() *gorp.DbMap {
	return sharedDBInstance
}

func New(playerId string, ctx interface{}) *MemStore {
	e := &MemStore{
		playerId:    playerId,
		store:       make(Store),
		storeStatus: make(StoreStatus),
		dataLoaded:  make(map[string]bool),
		Db:          GetSharedDBInstance(),
		Ctx:         ctx,
	}
	return e
}

func RegisterDataLoader(modelName string, loader dataLoader) {
	dataLoaderMap[modelName] = loader
}

func (e *MemStore) EnsureDataLoaded(modelName string) {
	if loaded, ok := e.dataLoaded[modelName]; !ok || !loaded {
		handler, ok := dataLoaderMap[modelName]
		if ok {
			logger.INFO("EnsureDataLoaded: ", e.playerId, " modelName: ", modelName)
			handler(e.playerId, e)
			e.dataLoaded[modelName] = true
		}
	}
}

func (e *MemStore) Load(namespaces []string, key string, value interface{}) {
	ctx := e.makeCtx(namespaces)
	ctx[key] = value
	e.UpdateStatus(namespaces[len(namespaces)-1], key, STATUS_ORIGIN)
}

func (e *MemStore) Get(namespaces []string, key string) interface{} {
	e.EnsureDataLoaded(namespaces[len(namespaces)-1])
	if ctx := e.getCtx(namespaces); ctx != nil {
		return ctx[key]
	} else {
		return nil
	}
}

func (e *MemStore) Set(namespaces []string, key string, value interface{}) {
	e.EnsureDataLoaded(namespaces[len(namespaces)-1])
	ctx := e.makeCtx(namespaces)
	if ctx[key] == nil {
		e.UpdateStatus(namespaces[len(namespaces)-1], key, STATUS_CREATE)
	} else {
		e.UpdateStatus(namespaces[len(namespaces)-1], key, STATUS_UPDATE)
	}
	ctx[key] = value
}

func (e *MemStore) Del(namespaces []string, key string) {
	e.EnsureDataLoaded(namespaces[len(namespaces)-1])
	if ctx := e.getCtx(namespaces); ctx != nil {
		e.UpdateStatus(namespaces[len(namespaces)-1], key, STATUS_DELETE)
		delete(ctx, key)
	}
}

func (e *MemStore) Find(namespaces []string, filter Filter) interface{} {
	e.EnsureDataLoaded(namespaces[len(namespaces)-1])
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
	e.EnsureDataLoaded(namespaces[len(namespaces)-1])
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
	e.EnsureDataLoaded(namespaces[len(namespaces)-1])
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
func (e *MemStore) Persist(namespaces []string) error {
	var err error
	sqls := make([]string, 0)
	for tableName, tableCtx := range e.getCtx(namespaces) {
		statusMap, ok := e.tableStatus(tableName)
		logger.INFO("tableName: ", tableName, " ok: ", ok)
		if ok {
			sqls, err = genTableSqls(sqls, statusMap, tableCtx.(Store))
			if err != nil {
				return err
			}
		}
	}
	if len(sqls) > 0 {
		err := AddPersistTask(e.playerId, time.Now().Unix(), sqls)
		if err != nil {
			logger.ERR("AddPersitTask failed, player: ", e.playerId, " err: ", err)
			return err
		}
		e.cleanStatus()
	}
	return nil
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
	if modelStatus, ok := tableStatus[key]; ok {
		modelStatus.Current = status
	} else {
		tableStatus[key] = &ModelStatus{
			Origin:  STATUS_EMPTY,
			Current: status,
		}
	}
}

func (e *MemStore) SetOriginStatus(table string, key string) {
	tableStatus, ok := e.storeStatus[table]
	if !ok {
		tableStatus = make(TableStatus)
		e.storeStatus[table] = tableStatus
	}
	tableStatus[key] = &ModelStatus{
		Origin:  STATUS_ORIGIN,
		Current: STATUS_EMPTY,
	}
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

func genTableSqls(sqls []string, statusMap TableStatus, tableCtx Store) ([]string, error) {
	for uuid, modelStatus := range statusMap {
		model := tableCtx[uuid].(ModelInterface)
		var status int8
		if modelStatus.Origin == STATUS_ORIGIN {
			switch modelStatus.Current {
			case STATUS_EMPTY:
				continue
			case STATUS_UPDATE, STATUS_DELETE:
				status = modelStatus.Current
				break
			case STATUS_CREATE:
				return nil, errors.New("genTableSqls should not STATUS_CREATE!")
			}
		} else if modelStatus.Origin == STATUS_EMPTY {
			switch modelStatus.Current {
			case STATUS_DELETE:
				continue
			case STATUS_UPDATE, STATUS_CREATE:
				status = STATUS_CREATE
				break
			case STATUS_ORIGIN, STATUS_EMPTY:
				return nil, errors.New("genTableSqls should not STATUS_ORIGIN or STATUS_EMPTY!")
			}
		}
		query := model.SqlForRec(status)
		logger.INFO("genTableSqls: ", query)
		sqls = append(sqls, query)
	}
	return sqls, nil
}
