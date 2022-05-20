package handler

import (
	"net/http"
	
	"github.com/andy-zhangtao/shadow-rest/shadowsocks/util"
)

// GetVersion 获取当前版本信息
func GetVersion(w http.ResponseWriter, r *http.Request) {
	
	dv := "Dev Version: " + "r53+5M 947d446"
	rv := "  Release Version: 0.2"
	util.HTTPSuccess(w, dv+rv)
}
