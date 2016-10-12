package shadowsocks

import (
	"errors"
	"net"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-11
 * Time: 11:03
 * Listen类及其维护函数
 */

var listenMap map[string]net.Listener

// Listen 网络监听对象
type Listen struct {
	Port   string
	Listen net.Listener
}

// AddListen 添加网络对象
func AddListen(l Listen) {
	if len(listenMap) == 0 {
		listenMap = make(map[string]net.Listener)
	}

	listenMap[l.Port] = l.Listen
}

// KillListen 关闭指定网络链接
func KillListen(port string) error {
	l := listenMap[port]

	if l != nil {
		err := l.Close()
		if err != nil {
			return err
		}
		delete(listenMap, port)
		return nil
	}

	return errors.New("没有此端口信息")
}

// GetListen 获取指定或者所有的网络链接
func GetListen() []string {
	key := make([]string, 0, len(listenMap))

	for k := range listenMap {
		key = append(key, k)
	}

	return key
}

// IsExists 判断指定端口是否已经存在
func IsExists(port string) bool {
	if listenMap[port] == nil {
		return false
	}

	return true
}
