package natter

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ForwardSocket 是一个用于转发 TCP 流量的结构体
type ForwardSocket struct {
	active   bool         `comment:"是否激活"`
	buffSize int          `comment:"缓冲区大小"`
	LAddr    *net.TCPAddr `comment:"本地地址"`
}

// NewForwardSocket 创建一个 ForwardSocket 实例
func NewForwardSocket(lAddr *net.TCPAddr) *ForwardSocket {
	return &ForwardSocket{
		active:   false,
		buffSize: 8192,
		LAddr:    lAddr,
	}
}

// StartForward 启动一个 ForwardSocket 实例
func (fs *ForwardSocket) StartForward(forwardAddr *net.TCPAddr) error {
	if fs.LAddr.String() == forwardAddr.String() {
		return fmt.Errorf("转发地址与监听地址不能为同一地址 %s", fs.LAddr.String())
	}
	conn, err := net.Dial("tcp4", forwardAddr.String())
	if err != nil {
		return fmt.Errorf("转发地址无法联通: %s", err.Error())
	}
	c(conn)

	logger.Info(fmt.Sprintf("开始端口转发 %s 到 %s", fs.LAddr.String(), forwardAddr.String()))

	fs.active = true
	go fs.socketTCPListen(fs.LAddr.Port, forwardAddr)
	return nil
}

// socketTCPListen 启动一个 TCP 监听
func (fs *ForwardSocket) socketTCPListen(listenPort int, forwardAddr *net.TCPAddr) {
	listenerAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		logger.Error(fmt.Sprintf("解析监听地址失败: %s", err.Error()))
		return
	}
	listener, err := net.ListenTCP("tcp", listenerAddr)
	if err != nil {
		logger.Error(fmt.Sprintf("启动监听失败: %s", err.Error()))
		return
	}
	logger.Debug(fmt.Sprintf("开始监听端口: %s", listenerAddr.String()))

	defer c(listener)
	for fs.active {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(fmt.Sprintf("监听端口数据读取失败,退出监听: %s", err.Error()))
			return
		}
		logger.Debug(fmt.Sprintf("接收到新连接: %s", conn.RemoteAddr().String()))
		go fs.socketTCPForward(conn, forwardAddr)
	}
}

// socketTCPForward 转发 TCP 数据
func (fs *ForwardSocket) socketTCPForward(conn net.Conn, forwardAddr *net.TCPAddr) {
	defer c(conn)
	outboundConn, err := net.Dial("tcp", forwardAddr.String())
	if err != nil {
		logger.Error(fmt.Sprintf("转发端口连接失败: %s", err.Error()))
		return
	}
	defer c(outboundConn)
	var wg sync.WaitGroup
	wg.Add(2)
	go fs.copyData(conn, outboundConn, &wg)
	go fs.copyData(outboundConn, conn, &wg)
	wg.Wait()
}

// copyData 复制数据
func (fs *ForwardSocket) copyData(src net.Conn, dst net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	buf := make([]byte, fs.buffSize)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return
		}
		if n > 0 {
			_, err = dst.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}
}

// StopForward 停止转发
func (fs *ForwardSocket) StopForward() {
	fs.active = false
}

// KeepAlive 用于保持连接
func (fs *ForwardSocket) KeepAlive() {
	dialer := &net.Dialer{
		LocalAddr: fs.LAddr,
	}
	timer := time.NewTimer(5 * time.Second)

	for fs.active {
		logger.Debug("keepAlive")
		<-timer.C
		conn, err := dialer.Dial("tcp4", keepLiveSrv)
		logger.Debug("keepAlive")
		if err != nil {
			logger.Error(fmt.Sprintf("保活请求失败: %s\n", err.Error()))
			continue
		}
		// 设置超时时间 3s
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _ = conn.Write([]byte("HEAD /natter-keep-alive HTTP/1.1\r\n" +
			"Host: www.baidu.com\r\n" +
			"User-Agent: curl/8.0.0 (Natter)\r\n" +
			"Accept: */*\r\n" +
			"Connection: keep-alive\r\n" +
			"\r\n"))
		_, _ = io.ReadAll(conn)
		_ = conn.Close()
		timer.Reset(5 * time.Second)
	}
}
