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
package test

import (
	"github.com/mafei198/gos/world/cache_mgr"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/mysqldb"
	"github.com/mafei198/gos/goslib/redisdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
