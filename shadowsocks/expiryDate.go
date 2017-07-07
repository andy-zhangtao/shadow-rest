package shadowsocks

import (
	"errors"
	"time"

	"github.com/andy-zhangtao/shadow-rest/configure"

	"github.com/andy-zhangtao/golog"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-12
 * Time: 17:30
 * 处理链接有效期
 */

// IsExpiry 当前链接是否有效
func IsExpiry() {
	for {
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		t := time.NewTimer(next.Sub(now))
		select {
		case <-t.C:
			for p := range listenMap {
				l := listenMap[p]
				ie, err := isExpiry(l)
				if err != nil {
					golog.Debug(err.Error())
					l.listen.Close()
					delete(listenMap, p)
				} else {
					if ie {
						golog.Debug(l.Port, "被关闭", l.ExpiryDate)
						l.listen.Close()
						delete(listenMap, p)
						KillUserPass(l.Port)
						SendEmail(l.Port+" 将会被关闭. 过期时间为:"+l.ExpiryDate, l.Email)
					}
				}
			}
		}
	}
}

// isExpiry 当前链接是否有效
// True 失效
// False 有效
func isExpiry(l Listen) (bool, error) {
	curr := time.Now()
	expiry, err := time.Parse(TIMEFORMATE, l.ExpiryDate)
	if err != nil {
		return true, err
	}
	return expiry.Before(curr), nil
}

// SetExpiry 设置指定网络链接有效期 @port 指定网络链接端口 @d 调整后的失效日期，只能为YYYY-MM-DD格式
func SetExpiry(port string, d string) error {
	l := listenMap[port]
	if l.Port == "" {
		return errors.New(configure.PORTNOTEXIST)
	}

	l.ExpiryDate = d

	listenMap[port] = l

	return nil
}
