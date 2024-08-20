package main

import (
	"context"
	"flag"
	"fmt"
	"natter/natter"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var logger = natter.GetLogger()

func main() {
	// 捕获异常
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err.(error).Error())
		}
		// 任意键退出
		fmt.Println("按回车键退出...")
		var input string
		_, _ = fmt.Scanln(&input)
	}()
	// 读取 命令行 参数
	var addr string
	flag.StringVar(&addr, "f", "127.0.0.1:80", "绑定公网地址至指定地址")
	flag.Parse()
	forwardAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = natter.Start(ctx, forwardAddr)
	if err != nil {
		panic(err)
	}
	// 等待系统中断信号
	// 创建一个 channel 用于接收信号
	c := make(chan os.Signal, 1)
	// 注册信号处理器
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// 阻塞直到有信号传入
	<-c
	fmt.Println("程序退出...")
	cancel()
}
