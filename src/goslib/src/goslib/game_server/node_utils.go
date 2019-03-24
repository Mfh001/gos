package game_server

import (
	"encoding/json"
	"fmt"
	"goslib/logger"
	"io/ioutil"
	"net/http"
)

const (
	metadataHost       = "http://metadata.google.internal"
	externalIPEndpoint = "/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"
	internalIPEndpoint = "/computeMetadata/v1/instance/network-interfaces/0/ip"
	tagsEndpoint       = "/computeMetadata/v1/instance/tags"
)

var (
	instanceIP string
)

type Endpoint struct {
	Name    string
	Address string
	Tags    []string
}

func PrintIps() {

	var err error

	instanceIP, err = getInstanceExternalIP()
	if err != nil {
		logger.ERR(err)
	}
	logger.INFO("externalIP: ", instanceIP)

	instanceIP, err = getInstanceIP()
	if err != nil {
		logger.ERR(err)
	}
	logger.INFO("instanceIP: ", instanceIP)

	tags, err := getInstanceTags()
	if err != nil {
		logger.ERR(err)
	}
	logger.INFO("tags: ", tags)
}

func getInstanceExternalIP() (string, error) {
	return getInstanceIPFromMetadata(true)
}

func getInstanceIP() (string, error) {
	return getInstanceIPFromMetadata(false)
}

func getInstanceIPFromMetadata(external bool) (string, error) {
	endpoint := internalIPEndpoint
	if external {
		endpoint = externalIPEndpoint
	}

	u := fmt.Sprintf("%s/%s", metadataHost, endpoint)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error retrieving instance IP: %d", resp.StatusCode)
	}

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func getInstanceTags() ([]string, error) {
	return getInstanceTagsFromMetadata()
}

func getInstanceTagsFromMetadata() ([]string, error) {
	var tags []string

	u := fmt.Sprintf("%s/%s", metadataHost, tagsEndpoint)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return tags, err
	}
	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return tags, err
	}

	if resp.StatusCode != 200 {
		return tags, fmt.Errorf("error retrieving instance tags: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tags)
	if err != nil {
		return tags, err
	}

	return tags, nil
}
