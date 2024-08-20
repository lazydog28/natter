package natter

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Start 启动 Natter
func Start(ctx context.Context, forwardAddr *net.TCPAddr) (err error) {
	logger.Info("开始获取 STUN 映射")
	lAddr, rAddr, err := GetMapping(nil)
	if err != nil {
		logger.Error(fmt.Sprintf("获取 STUN 映射失败: %s\n", err.Error()))
		return
	}
	logger.Info(fmt.Sprintf("获取 STUN 映射成功:"))
	logger.Info(fmt.Sprintf("内网地址: %s", lAddr.String()))
	logger.Info(fmt.Sprintf("公网地址: %s", rAddr.String()))

	forward := NewForwardSocket(lAddr)
	err = forward.StartForward(forwardAddr)
	if err != nil {
		logger.Error(fmt.Sprintf("启动端口转发失败: %s\n", err.Error()))
		return
	}
	logger.Info(fmt.Sprintf("公网访问地址: %s", rAddr.String()))
	go func() {
		logger.Info("开始保持 Natter 活跃")
		for {
			timer := time.NewTimer(5 * time.Second)
			<-timer.C
			select {
			case <-ctx.Done():
				logger.Info("关闭 Natter")
				forward.StopForward()
				return
			default:
				forward.KeepAlive()
				if !TestPort(rAddr.Port) {
					logger.Error("公网端口不可达")
					forward.StopForward()
					err = Start(ctx, forwardAddr)
					if err != nil {
						logger.Error(fmt.Sprintf("重新获取 STUN 映射失败: %s\n", err.Error()))
					}
					return
				}
			}
			timer.Reset(5 * time.Second)
		}
	}()
	return
}
