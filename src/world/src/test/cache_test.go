package test

import (
	"cache_mgr"
	"gen/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"goslib/mysqldb"
	"goslib/redisdb"
	"time"
)

var _ = Describe("Cache", func() {
	_ = mysqldb.StartClient()
	redisdb.StartClient()
	cache_mgr.StartPersister()

	var cache *cache_mgr.CacheMgr
	BeforeEach(func() {
		cache = &cache_mgr.CacheMgr{}
	})

	It("should can take", func() {
		_, err := cache.Take(nil, &proto.TakeRequest{
			PlayerId: "fake_player_id",
		})
		Expect(err).To(BeNil())
	})

	It("should can return", func() {
		reply, err := cache.Return(nil, &proto.ReturnRequest{
			PlayerId: "fake_player_id",
			Version:  time.Now().Unix(),
			Data:     "fake_data",
		})
		Expect(err).To(BeNil())
		Expect(reply.Success).Should(BeEquivalentTo(true))
	})

	It("should can persist", func() {
		reply, err := cache.Persist(nil, &proto.PersistRequest{
			PlayerId: "fake_player_id",
			Version:  time.Now().Unix(),
			Data:     "fake_data",
		})
		Expect(err).To(BeNil())
		Expect(reply.Success).Should(BeEquivalentTo(true))
	})

	It("should can ensure persisted", func() {
		cache_mgr.EnsurePersistered()
	})
})
