package mysqldb

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"gosconf"
)

var sharedDBInstance *sql.DB

func StartClient() error {
	var db *sql.DB
	var err error
	if flag.Lookup("test.v") == nil {
		switch gosconf.START_TYPE {
		case gosconf.START_TYPE_ALL_IN_ONE:
			db, err = sql.Open("mysql", gosconf.MYSQL_DSN_ALL_IN_ONE)
			break
		case gosconf.START_TYPE_CLUSTER:
			db, err = sql.Open("mysql", gosconf.MYSQL_DSN_CLUSTER)
			break
		case gosconf.START_TYPE_K8S:
			db, err = sql.Open("mysql", gosconf.MYSQL_DSN_K8S)
			break
		}
	} else {
		db, err = sql.Open("mysql", gosconf.MYSQL_DSN_ALL_IN_ONE)
	}
	if err != nil {
		return err
	}
	sharedDBInstance = db
	return nil
}

func Instance() *sql.DB {
	return sharedDBInstance
}
