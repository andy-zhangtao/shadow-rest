package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/andy-zhangtao/shadow-rest/configure"

	ss "github.com/andy-zhangtao/shadow-rest/shadowsocks"

	"github.com/andy-zhangtao/Sandstorm"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-12
 * Time: 18:40
 * 处理关于链接有效期的网络请求
 */

// Expiry 有效期结构体
type Expiry struct {
	Port   string `json:"port"`
	Expiry string `json:"expiry"`
}

// SetExpiryHandler 设置网络有效期
func SetExpiryHandler(w http.ResponseWriter, r *http.Request) {
	conf := new(Expiry)
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Sandstorm.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(content, &conf)
	if err != nil {
		Sandstorm.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conf.Port == "" {
		Sandstorm.HTTPError(w, configure.NOPORT, http.StatusInternalServerError)
		return
	}

	err = ss.SetExpiry(conf.Port, conf.Expiry)
	if err != nil {
		Sandstorm.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Sandstorm.HTTPSuccess(w, "OK")
}
