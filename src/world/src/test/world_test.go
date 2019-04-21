package test

import (
	. "github.com/onsi/ginkgo"
	"goslib/mysqldb"
)

var _ = Describe("World", func() {
	mysqldb.StartClient()
	It("should startup", func() {
	})
})
