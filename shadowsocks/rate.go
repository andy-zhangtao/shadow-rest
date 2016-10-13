package shadowsocks

import "fmt"

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
