package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"account"
	"fmt"
	"redisDB"
)

var _ = Describe("AuthApp", func() {
	redisDB.Connect("localhost:6379", "", 0)

	It("should startup", func() {
		accountId := "fakeAccountId"
		password := "fakePassword"

		account.Delete(accountId)
		account.Create(accountId, password)

		user := account.Lookup(accountId)
		fmt.Println("Found user: ", user.Username)
		Expect(user).ToNot(BeNil())
	})
})
