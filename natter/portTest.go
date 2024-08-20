package natter

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// dialIPv4Context 是一个自定义的 DialContext 函数，用于指定 IPv4 地址
var dialIPv4Context = func(ctx context.Context, network, addr string) (net.Conn, error) {
	return net.Dial("tcp4", addr)
}

// 创建一个自定义的 Transport 以使用 dialIPv4Context
var tr = &http.Transport{
	DialContext: dialIPv4Context,
}

// testIfConfigCo 通过 ifconfig.co 确定端口是否可达
func testIfConfigCo(port int) bool {
	url := fmt.Sprintf("https://ifconfig.co/port/%d", port)
	method := "GET"
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	res, err := client.Do(req)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	defer c(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	var reachable bool
	if strings.Contains(string(body), "true") {
		reachable = true
	}
	logger.Debug(fmt.Sprintf("ifconfig.co 端口 %d 返回: %v", port, reachable))
	return reachable
}

// testTransmission 通过 Transmission 端口检查服务确定端口是否可达
func testTransmission(port int) bool {
	url := fmt.Sprintf("https://portcheck.transmissionbt.com/%d", port)
	method := "GET"
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	res, err := client.Do(req)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	defer c(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	bodyStr := string(body)
	// 移除前后空白字符
	bodyStr = strings.TrimSpace(bodyStr)
	logger.Debug(fmt.Sprintf("Transmission 端口 %d 检查服务返回: %s", port, bodyStr))
	return bodyStr == "1"
}

func TestPort(port int) bool {
	return testIfConfigCo(port) || testTransmission(port)
}
