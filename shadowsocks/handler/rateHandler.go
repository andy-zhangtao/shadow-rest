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
	ss "shadow-rest/shadowsocks"

	"github.com/andy-zhangtao/Sandstorm"
)

// GetRateHandler 获取当前所有的端口流量
func GetRateHandler(w http.ResponseWriter, r *http.Request) {
	keys := ss.GetRate()

	content, _ := json.Marshal(keys)

	Sandstorm.HTTPSuccess(w, string(content))
}
