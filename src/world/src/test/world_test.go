package test

import (
	"gen/register"
	. "github.com/onsi/ginkgo"
	"goslib/mysqldb"
)

var _ = Describe("World", func() {
	mysqldb.StartClient()
	register.RegisterTables(mysqldb.Instance())

	It("should startup", func() {
	})
})
