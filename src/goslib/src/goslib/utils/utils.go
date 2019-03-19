package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"goslib/logger"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

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
