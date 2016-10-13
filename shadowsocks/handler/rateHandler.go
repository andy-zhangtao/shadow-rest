package handler

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-11
 * Time: 18:37
 * 处理关于Rate的网络请求
 */

import (
	"encoding/json"
	"net/http"

	ss "github.com/andy-zhangtao/shadow-rest/shadowsocks"

	"github.com/andy-zhangtao/Sandstorm"
)

// GetInfoHandler 获取当前所有的端口信息
func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	keys := ss.GetRate()

	content, _ := json.Marshal(keys)

	Sandstorm.HTTPSuccess(w, string(content))
}

// GetRateHandler 获取当前所有端口流量数据
func GetRateHandler(w http.ResponseWriter, r *http.Request) {
	keys := ss.GetRate()
	rate := make([]*ss.Rate, len(keys))

	for i, k := range keys {
		r := &ss.Rate{
			Port: k.Port,
			Rate: ss.ConvertRate(k.Rate),
		}
		rate[i] = r
	}

	content, _ := json.Marshal(rate)

	Sandstorm.HTTPSuccess(w, string(content))
}
