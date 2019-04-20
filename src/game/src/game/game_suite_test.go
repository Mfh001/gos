package main

import (
	"goslib/redisdb"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGame(t *testing.T) {
	redisdb.StartClient()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Game Suite")
}
