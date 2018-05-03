package gslib

import (
	"app/register/tables"
	"fmt"
	"goslib/broadcast"
	"goslib/gen_server"
	"goslib/memstore"
	"time"
)

func Run() {
	//defer func() {
	//	if x := recover(); x != nil {
	//		fmt.Println("caught panic in main()", x)
	//	}
	//}()

	go SysRoutine()

	time.Sleep(1 * time.Second)

	fmt.Println("Server Started!")

	// Init DB Connections
	memstore.InitDB()
	memstore.StartDBPersister()
	tables.RegisterTables(memstore.GetSharedDBInstance())

	// Start broadcast server
	gen_server.Start(BROADCAST_SERVER_ID, new(broadcast.Broadcast))
}
