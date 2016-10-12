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

var listenMap map[string]Listen

// Listen 网络监听对象
type Listen struct {
	Port   string `json:"port"`
	Rate   int    `json:"rate"`
	Listen net.Listener
}

// AddListen 添加网络对象
func AddListen(l Listen) {
	if len(listenMap) == 0 {
		listenMap = make(map[string]Listen)
	}

	listenMap[l.Port] = l
}

// KillListen 关闭指定网络链接
func KillListen(port string) error {
	l := listenMap[port]

	if l.Listen != nil {
		err := l.Listen.Close()
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
	l := listenMap[port]
	if l.Listen == nil {
		return false
	}

	return true
}

// AddRate 按照端口统计流量
func AddRate(r Listen) {
	l := listenMap[r.Port]
	l.Rate = l.Rate + r.Rate
	listenMap[r.Port] = l
}

// GetPortRate 获取指定端口流量
func GetPortRate(port string) *Listen {
	l := listenMap[port]
	r := &Listen{
		Port: port,
		Rate: l.Rate,
	}

	return r
}

// GetRate 获取所有端口流量
func GetRate() []*Listen {
	r := make([]*Listen, 0, len(listenMap))
	for p := range listenMap {
		l := listenMap[p]
		rate := &Listen{
			Port: p,
			Rate: l.Rate,
		}

		r = append(r, rate)
	}

	return r
}

// ClearPortRate 端口流量清零
func ClearPortRate(port string) {
	l := listenMap[port]
	l.Rate = 0
	listenMap[port] = l
}
