package store

import (
	"fmt"
	"strings"
)

type Filter func(elem interface{}) bool

type packet struct {
	action     byte
	namespaces []string
	key        string
	value      interface{}
	cb         Filter
}

type Store map[string]interface{}

type Ets struct {
	channel_in  chan *packet
	channel_out chan interface{}
	store       Store
}

const (
	GET    = 1
	LOAD   = 2
	SET    = 3
	DEL    = 4
	FIND   = 5
	SELECT = 6
)

const (
	STATUS_ORIGIN = 1
	STATUS_CREATE = 2
	STATUS_UPDATE = 3
	STATUS_DELETE = 4
)

var sharedInstance *Ets

func InitSharedInstance() {
	sharedInstance = New()
}

func New() *Ets {
	e := &Ets{
		channel_in:  make(chan *packet),
		channel_out: make(chan interface{}),
		store:       make(Store),
	}
	go e.loop()
	return e
}

func Get(namespaces []string, key string) interface{} {
	return sharedInstance.Get(namespaces, key)
}
func Load(namespaces []string, key string, value interface{}) {
	sharedInstance.Load(namespaces, key, value)
}
func Set(namespaces []string, key string, value interface{}) {
	sharedInstance.Set(namespaces, key, value)
}
func Del(namespaces []string, key string) {
	sharedInstance.Del(namespaces, key)
}
func Find(namespaces []string, filter Filter) interface{} {
	return sharedInstance.Find(namespaces, filter)
}
func Select(namespaces []string, filter Filter) interface{} {
	return sharedInstance.Select(namespaces, filter)
}

func (e *Ets) Get(namespaces []string, key string) interface{} {
	e.channel_in <- &packet{action: GET, namespaces: namespaces, key: key}
	value := <-e.channel_out
	return value
}

func (e *Ets) Load(namespaces []string, key string, value interface{}) {
	e.channel_in <- &packet{action: LOAD, namespaces: namespaces, key: key, value: value}
}

func (e *Ets) Set(namespaces []string, key string, value interface{}) {
	e.channel_in <- &packet{action: SET, namespaces: namespaces, key: key, value: value}
}

func (e *Ets) Del(namespaces []string, key string) {
	e.channel_in <- &packet{action: DEL, namespaces: namespaces, key: key}
}

func (e *Ets) Find(namespaces []string, filter Filter) interface{} {
	e.channel_in <- &packet{action: FIND, namespaces: namespaces, cb: filter}
	value := <-e.channel_out
	return value
}

func (e *Ets) Select(namespaces []string, filter Filter) interface{} {
	e.channel_in <- &packet{action: SELECT, namespaces: namespaces, cb: filter}
	value := <-e.channel_out
	return value
}

func (e *Ets) getCtx(namespaces []string) Store {
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
	// fmt.Println("getCtx: ", ctx)
	return ctx
}

func (e *Ets) makeCtx(namespaces []string) Store {
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

func (e *Ets) loop() {
	for {
		data := <-e.channel_in
		switch data.action {
		case GET:
			ctx := e.getCtx(data.namespaces)
			if ctx != nil {
				e.channel_out <- ctx[data.key]
			} else {
				e.channel_out <- nil
			}
		case LOAD:
			ctx := e.makeCtx(data.namespaces)
			ctx[data.key] = data.value
		case SET:
			ctx := e.makeCtx(data.namespaces)
			if ctx[data.key] == nil {
				e.updateStatus(data.namespaces, data.key, STATUS_CREATE)
			} else {
				e.updateStatus(data.namespaces, data.key, STATUS_UPDATE)
			}
			ctx[data.key] = data.value
		case DEL:
			if ctx := e.getCtx(data.namespaces); ctx != nil {
				e.updateStatus(data.namespaces, data.key, STATUS_DELETE)
				delete(ctx, data.key)
			}
		case FIND:
			if ctx := e.getCtx(data.namespaces); ctx != nil {
				for _, v := range ctx {
					if data.cb(v) {
						e.channel_out <- v
						break
					}
				}
			} else {
				e.channel_out <- nil
			}
		case SELECT:
			ctx := e.getCtx(data.namespaces)
			if ctx != nil {
				var elems []interface{}
				for _, v := range ctx {
					if data.cb(v) {
						elems = append(elems, v)
					}
				}
				e.channel_out <- elems
			} else {
				e.channel_out <- nil
			}
		}
	}
}

func (e *Ets) updateStatus(namespaces []string, key string, status int) {
	statusKey := getStatusKey(namespaces)
	ctx, ok := e.store[statusKey]
	if !ok {
		ctx = make(Store)
		e.store[statusKey] = ctx
	}
	ctx.(Store)[key] = status
}

func (e *Ets) allStatus(namespaces []string) Store {
	return e.store[getStatusKey(namespaces)].(Store)
}

func (e *Ets) cleanStatus(namespaces []string) {
	delete(e.store, getStatusKey(namespaces))
}

func getStatusKey(namespaces []string) string {
	return strings.Join(append(namespaces, "status"), "_")
}

func (e *Ets) persistAll(namespaces []string) {
	var sqls []string
	for tableName, tableCtx := range e.getCtx(namespaces) {
		status := e.allStatus([]string{namespaces[1], tableName})
		sqls = e.generateSql(sqls, tableName, status, tableCtx.(Store))
	}
	strings.Join(sqls, ";")
	// db.Execute(sql)
}

func (e *Ets) generateSql(sqls []string, tableName string, status Store, tableCtx Store) []string {
	for k, v := range status {
		var sql string
		switch v.(int) {
		case STATUS_UPDATE:
		case STATUS_DELETE:
			sql = fmt.Sprintf("DELETE FROM %s WHERE `uuid`=`%s`", tableName, k)
		case STATUS_CREATE:
		}
		sqls = append(sqls, sql)
	}
	return sqls
}
