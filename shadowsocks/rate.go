package shadowsocks

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-11
 * Time: 18:03
 * 管理每个端口OUTPUT流量,以字节为单位
 */

// Rate 流量类
// type Rate struct {
// 	Port string `json:"port"`
// 	Rate int    `json:"rate"`
// }

// var rateMap map[string]int

// AddRate 按照端口统计流量
// func AddRate(r Rate) {
// 	if len(rateMap) == 0 {
// 		rateMap = make(map[string]int)
// 	}

// 	rateMap[r.Port] = rateMap[r.Port] + r.Rate
// }

// GetPortRate 获取指定端口流量
// func GetPortRate(port string) *Rate {
// 	r := &Rate{
// 		Port: port,
// 		Rate: rateMap[port],
// 	}

// 	return r
// }

// GetRate 获取所有端口流量
// func GetRate() []*Rate {
// 	r := make([]*Rate, 0, len(rateMap))
// 	for p := range rateMap {
// 		rate := &Rate{
// 			Port: p,
// 			Rate: rateMap[p],
// 		}

// 		r = append(r, rate)
// 	}

// 	return r
// }

// ClearPortRate 端口流量清零
// func ClearPortRate(port string) {
// 	rateMap[port] = 0
// }
