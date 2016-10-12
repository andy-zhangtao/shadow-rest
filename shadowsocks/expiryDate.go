package shadowsocks

import (
	"errors"
	"shadow-rest/configure"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-12
 * Time: 17:30
 * 处理链接有效期
 */

// IsExpiry 当前链接是否有效
// True 失效
// False 有效
// func IsExpiry(l *Listen) bool {
//
// }

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
