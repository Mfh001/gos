package main

import (
	"gen/register"
	. "github.com/onsi/ginkgo"
	"goslib/memstore"
)

var _ = Describe("World", func() {
	memstore.StartDB()
	memstore.StartDBPersister()
	register.RegisterTables(memstore.GetSharedDBInstance())

	It("should startup", func() {
	})
})
