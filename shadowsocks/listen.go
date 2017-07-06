package shadowsocks

import (
	"errors"
	"net"
	"os"

	"github.com/andy-zhangtao/shadow-rest/shadowsocks/db"

	"github.com/andy-zhangtao/shadow-rest/configure"
	"github.com/andy-zhangtao/shadow-rest/shadowsocks/util"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-11
 * Time: 11:03
 * Listen类及其维护函数
 */

var listenMap map[string]Listen
var PassMap map[string]UserPass

// GlobaIP 全局监听IP
var GlobaIP string

const (
	// TIMEFORMATE 默认时间格式
	TIMEFORMATE = "2006-01-02"
)

// Listen 网络监听对象
type Listen struct {
	Port       string `json:"port"`
	Rate       int    `json:"rate"`
	ExpiryDate string `json:"expiry_date"`
	RateLimit  int    `json:"rate_limit"`
	Email      string `json:"email"` //16-12-14 新增Email属性
	listen     net.Listener
}

// ListenBak 网络备份
type ListenBak struct {
	Lb []Listen `json:"lb"`
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

	if l.listen != nil {
		err := l.listen.Close()
		if err != nil {
			return err
		}
		delete(listenMap, port)

		KillUserPass(port)

		return nil
	}

	return errors.New(configure.PORTNOTEXIST)
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
	if l.listen == nil {
		return false
	}

	return true
}

// AddRate 按照端口统计流量
func AddRate(r Listen) {
	l := listenMap[r.Port]
	l.Rate = l.Rate + r.Rate
	// 这里应该为listenMap添加同步锁，但考虑到效率所以再次判断一次
	tl := listenMap[r.Port]
	if tl.Port != "" {
		listenMap[r.Port] = l
	} else {
		if tl.listen != nil {
			tl.listen.Close()
		}
	}
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
func GetRate() ([]User, error) {
	if mongoSession == nil {
		mongoSession = db.GetMongo()
	}

	u := mongoSession.DB(os.Getenv(util.MONGODB)).C("user")

	var user []User

	err := u.Find(nil).Sort("+id").All(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ClearPortRate 端口流量清零
func ClearPortRate(port string) {
	l := listenMap[port]
	l.Rate = 0
	listenMap[port] = l
}
