package mysqldb

import (
	"database/sql"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
)

var sharedDBInstance *gorp.DbMap

func StartClient() error {
	db, err := sql.Open("mysql", "root:euQRdwMgb1@tcp(single-mysql.default.svc.cluster.local)/gos_server_development")
	if err != nil {
		return err
	}
	sharedDBInstance = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	return nil
}

func Instance() *gorp.DbMap {
	return sharedDBInstance
}
