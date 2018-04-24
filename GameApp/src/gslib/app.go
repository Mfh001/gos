package gslib

import (
	"fmt"
	"goslib/gen_server"
	"goslib/memStore"
	"time"
	"app/register/tables"
	"goslib/broadcast"
)

func Run() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("caught panic in main()", x)
		}
	}()

	go SysRoutine()

	time.Sleep(1 * time.Second)

	fmt.Println("Server Started!")

	// Init DB Connections
	memStore.InitDB()
	tables.RegisterTables(memStore.GetSharedDBInstance())

	// Start broadcast server
	gen_server.Start(BROADCAST_SERVER_ID, new(broadcast.Broadcast))
}

