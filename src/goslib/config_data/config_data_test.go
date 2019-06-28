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

package config_data

import (
	"github.com/mafei198/gos/goslib/gen/gd"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/redisdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	redisdb.StartClient()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Data")
}

var _ = Describe("pipelining", func() {
	It("Should get configdata from redis", func() {
		data, err := getConfigData()
		Expect(err).To(BeNil())
		Expect(len(data)).NotTo(BeZero())
	})

	It("Should can load config", func() {
		data, _ := getConfigData()
		gd.LoadConfigs(data)
	})

	It("Should can watch update", func() {
		Load()
		time.Sleep(2 * time.Second)
		channels, err := redisdb.Instance().PubSubChannels(gosconf.CONFIG_RELOAD_CHANNEL).Result()
		Expect(err).To(BeNil())
		Expect(channels).To(HaveLen(1))

		_, err = redisdb.Instance().Publish(gosconf.CONFIG_RELOAD_CHANNEL, "update").Result()
		Expect(err).To(BeNil())
	})
})
