package config_data

import (
	"gen/gd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gosconf"
	"goslib/redisdb"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Data")
}

var _ = Describe("pipelining", func() {
	redisdb.StartClient()

	It("Should get configdata from redis", func() {
		data, err := getConfigData()
		Expect(err).To(BeNil())
		Expect(len(data)).NotTo(BeZero())
	})

	It("Should can load config", func() {
		data, _ := getConfigData()
		gd.LoadConfigs(data)
		gd.GetGlobal()
		Expect(len(gd.BgmsIns.GetList())).NotTo(BeZero())
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
