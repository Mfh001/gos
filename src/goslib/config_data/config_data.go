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
	"bytes"
	"compress/gzip"
	"github.com/json-iterator/go"
	"github.com/mafei198/gos/goslib/gen/gd"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"io"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var AfterConfigLoaded func()

func Load() {
	go func() {
		for {
			if data, err := getConfigData(); err == nil {
				gd.LoadConfigs(data)
				if AfterConfigLoaded != nil {
					AfterConfigLoaded()
				}
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
					if AfterConfigLoaded != nil {
						AfterConfigLoaded()
					}
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

	reader := bytes.NewBuffer([]byte(content))
	readCloser, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	var replyData bytes.Buffer
	_, err = io.Copy(&replyData, readCloser)
	_ = readCloser.Close()
	if err != nil {
		return nil, err
	}

	//err = json.UnmarshalFromString(content, &data)
	err = json.Unmarshal(replyData.Bytes(), &data)
	if err != nil {
		logger.ERR("loadConfigs unmarshal content failed: ", err)
		return
	}
	return
}
