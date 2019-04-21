package config_data

import (
	"gen/gd"
	"github.com/json-iterator/go"
	"gosconf"
	"goslib/logger"
	"goslib/redisdb"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Load() {
	go func() {
		for {
			if data, err := getConfigData(); err == nil {
				gd.LoadConfigs(data)
				watchReload()
				break
			} else {
				time.Sleep(gosconf.RETRY_BACKOFF)
			}
		}
	}()
}

func watchReload() {
	for {
		pubsub := redisdb.Instance().Subscribe(gosconf.CONFIG_RELOAD_CHANNEL)
		if _, err := pubsub.Receive(); err != nil {
			time.Sleep(gosconf.RETRY_BACKOFF)
			continue
		}
		ch := pubsub.Channel()
		for msg := range ch {
			println("watchReload: ", msg.Payload)
			for {
				if data, err := getConfigData(); err == nil {
					gd.LoadConfigs(data)
					break
				} else {
					time.Sleep(gosconf.RETRY_BACKOFF)
				}
			}
		}
	}
}

func getConfigData() (data map[string]string, err error) {
	content, err := redisdb.Instance().Get(gosconf.CONFIG_GET_KEY).Result()
	if err != nil {
		logger.ERR("loadConfigs failed: ", err)
		return
	}

	err = json.UnmarshalFromString(content, &data)
	if err != nil {
		logger.ERR("loadConfigs unmarshal content failed: ", err)
		return
	}
	return
}
