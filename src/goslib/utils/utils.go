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
package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/mafei198/gos/goslib/logger"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var Chars = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
	"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z",
}
var CharsCount = len(Chars)

func RandChar() string {
	return Chars[rand.Intn(CharsCount)]
}

func RecoverPanic(where string) {
	if x := recover(); x != nil {
		stack := string(debug.Stack())
		errorMsg := format("caught panic in ", where, x, stack)
		raven.CaptureMessage(errorMsg, map[string]string{"category": "panic"})
		logger.ERRDirect(errorMsg)
	}
}

func format(v ...interface{}) string {
	return fmt.Sprintf("%v \033[0m\n", strings.TrimRight(fmt.Sprintln(v...), "\n"))
}

func GenId(list []string) string {
	content := strings.Join(list, "-")
	hasher := md5.New()
	hasher.Write([]byte(content))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				// Don't use random as we have a real address
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}

func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.ERR(err)
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.To4().String(), nil
}

func GetPublicIP() (string, error) {
	response, err := http.Get("http://ipinfo.io/ip")
	if err != nil {
		logger.ERR("GetPublicIP failed: ", err)
		return "", err
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.ERR("GetPublicIP failed: ", err)
		return "", err
	}
	return strings.Trim(string(content), "\n"), nil
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		logger.ERR("GetLocalIP failed: ", err)
		return "", err
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}
	return "", nil
}

func GetHostname() (string, error) {
	return os.Hostname()
}

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func AbsInt32(v int32) int32 {
	if v < 0 {
		return -v
	} else {
		return v
	}
}

func FloorToInt32(v float32) int32 {
	return int32(math.Floor(float64(v)))
}

func CeilToInt32(v float32) int32 {
	return int32(math.Ceil(float64(v)))
}

func CeilToInt64(v float32) int64 {
	return int64(math.Ceil(float64(v)))
}

var DAY_SECONDS int64 = 86400

func NDaysAgo(n int32) int64 {
	return time.Now().Unix() - int64(n)*DAY_SECONDS
}

func StructToStr(ins interface{}) string {
	return fmt.Sprintf("%+v\n", ins)
}

const Backoff = 1
const BackoffRatio = 2
const MaxBackoff = 3

func Retry(maxRetry int, handler func() error) error {
	backoff := Backoff
	for i := 0; ; i++ {
		err := handler()
		if err == nil || i == maxRetry {
			return err
		}

		time.Sleep(time.Duration(backoff) * time.Second)
		if i < MaxBackoff {
			backoff *= BackoffRatio
		}
	}
}
