package store

import (
	_ "database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	// "log"
	"reflect"
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
	db          *sqlx.DB
}

const (
	GET     = 1
	LOAD    = 2
	SET     = 3
	DEL     = 4
	FIND    = 5
	SELECT  = 6
	PERSIST = 7
)

const (
	STATUS_ORIGIN = 1
	STATUS_CREATE = 2
	STATUS_UPDATE = 3
	STATUS_DELETE = 4
)

var sharedInstance *Ets = nil

func InitSharedInstance() {
	if sharedInstance == nil {
		sharedInstance = New()
	}
}

type Equip struct {
	Uuid    string `db:"uuid"`
	UserId  string `db:"user_id"`
	Level   int    `db:"level"`
	ConfId  int    `db:"conf_id"`
	Evolves string `db:"evolves"`
	Equiped string `db:"equiped"`
	Exp     int    `db:"exp"`
}

func Test() {
	equip := &Equip{}
	sharedInstance.db.Select(&equip, "select * from equips where uuid='54A3927E2B89780A1491F441'")
	fmt.Println("Store Test:", equip)

	key := "54A3927E2B89780A1491F441"
	namespaces := []string{"54A3927E2B89780A1491F43C", "equips"}
	Load(namespaces, key, equip)

	fmt.Println("Get: ", Get(namespaces, key))

	equip.Level = 10
	Set(namespaces, key, equip)
	Del(namespaces, "1")
	Persist([]string{"54A3927E2B89780A1491F43C"})
}

func New() *Ets {
	// db := sqlx.MustConnect("mysql", "user=root dbname=game_server_development sslmode=disable")
	db := sqlx.MustConnect("mysql", "root:@/game_server_development")

	e := &Ets{
		channel_in:  make(chan *packet),
		channel_out: make(chan interface{}),
		store:       make(Store),
		db:          db,
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
func Persist(namespaces []string) {
	sharedInstance.Persist(namespaces)
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

func (e *Ets) Persist(namespaces []string) {
	e.channel_in <- &packet{action: PERSIST, namespaces: namespaces}
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
		case PERSIST:
			e.persistAll(data.namespaces)
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

func (e *Ets) getStatus(namespaces []string, key string) int {
	statusKey := getStatusKey(namespaces)
	ctx, ok := e.store[statusKey]
	if !ok {
		return STATUS_ORIGIN
	} else {
		return ctx.(Store)[key].(int)
	}
}

func (e *Ets) allStatus(namespaces []string) Store {
	ctx := e.store[getStatusKey(namespaces)]
	if ctx != nil {
		return ctx.(Store)
	} else {
		return nil
	}
}

func (e *Ets) cleanStatus(namespaces []string) {
	delete(e.store, getStatusKey(namespaces))
}

func getStatusKey(namespaces []string) string {
	return strings.Join(append(namespaces, "status"), "_")
}

func (e *Ets) persistAll(namespaces []string) {
	tx := e.db.MustBegin()
	for tableName, tableCtx := range e.getCtx(namespaces) {
		status := e.allStatus([]string{namespaces[0], tableName})
		e.executeSql(tx, tableName, status, tableCtx.(Store))
	}
	err := tx.Commit()
	if err != nil {
		panic(err.Error())
	}
	e.cleanStatus(namespaces)
}

func (e *Ets) executeSql(tx *sqlx.Tx, tableName string, status Store, tableCtx Store) {
	for k, v := range status {
		switch v.(int) {
		case STATUS_UPDATE:
			fields, values := joinFieldsAndValuesForUpdate(tableCtx[k])
			tx.MustExec(fmt.Sprintf("UPDATE `%s` SET %s WHERE `uuid` = '%s'", tableName, fields, k), values...)
		case STATUS_DELETE:
			tx.MustExec(fmt.Sprintf("DELETE FROM `%s` WHERE `uuid`='%s'", tableName, k))
		case STATUS_CREATE:
			fields, valuesPlaceHolder, values := joinFieldsAndValuesForCreate(tableCtx[k])
			tx.MustExec(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, fields, valuesPlaceHolder), values...)
		}
	}
}

func joinFieldsAndValuesForUpdate(s interface{}) (string, []interface{}) {
	st := reflect.ValueOf(s)
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface:
		st = st.Elem()
	}
	var fields []string
	var values []interface{}
	for i := 0; i < st.NumField(); i++ {
		fields = append(fields, fmt.Sprintf("`%s` = ?", st.Type().Field(i).Name))
		values = append(values, fmt.Sprint(st.Field(i).Interface()))
	}
	return strings.Join(fields, ", "), values
}

func joinFieldsAndValuesForCreate(s interface{}) (string, string, []interface{}) {
	st := reflect.ValueOf(s)
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface:
		st = st.Elem()
	}
	var fields []string
	var values []interface{}
	var valuesPlaceHolder []string
	for i := 0; i < st.NumField(); i++ {
		fields = append(fields, fmt.Sprintf("`%s`", st.Type().Field(i).Name))
		values = append(values, fmt.Sprint(st.Field(i).Interface()))
		valuesPlaceHolder = append(valuesPlaceHolder, "?")
	}
	return strings.Join(fields, ", "), strings.Join(valuesPlaceHolder, ", "), values
}
