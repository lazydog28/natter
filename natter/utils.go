package natter

import (
	"fmt"
	"io"
	"strings"
)

// c 用于关闭一个 io.Closer 接口
func c(close io.Closer) {
	if err := close.Close(); err != nil {
		// 判断 err 是否为 连接已关闭
		if strings.Contains(err.Error(), "use of closed network connection") {
			return
		}
		logger.Error(fmt.Sprintf("关闭连接失败: %s\n", err.Error()))
	}
}
