package shadowsocks

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/andy-zhangtao/shadow-rest/configure"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-13
 * Time: 10:10
 * 处理流量相关请求
 */

// Rate 流量数据 每次都从Listen中重新获取数据
type Rate struct {
	Port string `json:"port"`
	Rate string `json:"rate"`
}

// ConvertRate 流量转换。 Btye --> KB --> MB --> GB
func ConvertRate(rate int) string {
	if rate == 0 {
		return "0KB"
	}

	count := 1
	f := float64(rate)
	f = f / 1024

	for {

		if isOK(f) {
			break
		}
		count++
		f = f / 1024
	}

	switch count {
	case 1:
		return fmt.Sprintf("%0.3f KB", f)
	case 2:
		return fmt.Sprintf("%0.3f MB", f)
	case 3:
		return fmt.Sprintf("%0.3f GB", f)
	default:
		return fmt.Sprintf("%0.3f Btye", f)
	}
}

// isOK 判断当前流量值是否大于1024. 当大于1024时，返回false，继续换算。反之停止换算
func isOK(i float64) bool {
	return i <= 1024
}

// SetRate 设置网络链接流量上限
func SetRate(port, rate string) error {
	l := listenMap[port]
	if l.Port == "" {
		return errors.New(configure.PORTNOTEXIST)
	}

	rate = strings.TrimSpace(rate)
	tr := []byte(rate)
	r := tr[len(tr)-2:]
	v := tr[:len(tr)-2]
	vv := 0
	switch strings.ToLower(string(r)) {
	case "kb":
		rateVale, err := strconv.Atoi(string(v))
		if err != nil {
			return err
		}
		vv = rateVale * 1024

	case "mb":
		rateVale, err := strconv.Atoi(string(v))
		if err != nil {
			return err
		}
		vv = rateVale * 1024 * 1024

	case "gb":
		rateVale, err := strconv.Atoi(string(v))
		if err != nil {
			return err
		}
		vv = rateVale * 1024 * 1024 * 1024
	default:
		return errors.New(configure.INVALIDRATE)
	}

	l.RateLimit = vv
	listenMap[port] = l
	return nil
}

// IsAboveRate 判断当前每个链接是否超过指定流量
func IsAboveRate() {
	for {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				for p := range listenMap {
					l := listenMap[p]
					// 0 表示无限制
					if l.RateLimit != 0 {
						if l.Rate >= l.RateLimit {
							log.Println(l.Port, "流量超限需要被关闭")
							l.listen.Close()
							delete(listenMap, p)
							KillUserPass(l.Port)
							log.Println(l.Port, "被关闭", l.Rate, l.RateLimit)
							SendEmail(l.Port+" Will be closed. Curreny Rate:"+strconv.Itoa(l.Rate)+" Max Rate:"+strconv.Itoa(l.RateLimit), l.Email)
						}
					}

				}
			}
		}
	}
}
