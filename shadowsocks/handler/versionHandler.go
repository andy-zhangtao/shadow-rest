package handler

import (
	"net/http"

	"github.com/andy-zhangtao/Sandstorm"
)

// GetVersion 获取当前版本信息
func GetVersion(w http.ResponseWriter, r *http.Request) {

	dv := "Dev Version: " + "r44M f68e5b3"
	rv := "  Release Version: 0.1"
	Sandstorm.HTTPSuccess(w, dv+rv)
}
