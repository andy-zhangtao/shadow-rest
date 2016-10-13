package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/andy-zhangtao/shadow-rest/configure"
	ss "github.com/andy-zhangtao/shadow-rest/shadowsocks"

	"github.com/andy-zhangtao/Sandstorm"
	"github.com/gorilla/mux"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-11
 * Time: 11:03
 * 处理关于Listen的网络请求
 */

// RestartL 重启参数
type RestartL struct {
	Port     string `json:"port"`
	Password string `json:"password"`
	Method   string `json:"method,omitempty"`
	Auth     bool   `json:"auth,omitempty"`
}

// GetListenHandler 获取当前所有的网络链接
func GetListenHandler(w http.ResponseWriter, r *http.Request) {
	keys := ss.GetListen()

	Sandstorm.HTTPSuccess(w, strings.Join(keys, ","))
}

// DeleteListenHandler 删除指定网络链接
func DeleteListenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	port := vars["ports"]

	err := ss.KillListen(port)
	if err != nil {
		Sandstorm.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	Sandstorm.HTTPSuccess(w, "OK")
}

// RestartListenHandler 重启指定网络链接
func RestartListenHandler(w http.ResponseWriter, r *http.Request) {
	conf := new(RestartL)
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

	if ss.IsExists(conf.Port) {
		Sandstorm.HTTPError(w, configure.HASPORT, http.StatusInternalServerError)
		return
	}

	if conf.Password == "" {
		conf.Password = configure.DEFAULTPASSWD
	}

	if conf.Method == "" {
		conf.Method = configure.DEFAULTMETHOD
	}

	go ss.Run(conf.Port, conf.Password, conf.Method, conf.Auth)
}
