/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package mysqldb

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mafei198/gos/goslib/gosconf"
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
