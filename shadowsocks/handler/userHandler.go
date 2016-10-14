package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/andy-zhangtao/Sandstorm"

	ss "github.com/andy-zhangtao/shadow-rest/shadowsocks"
)

// CreateUserHandler 创建新用户
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	conf := new(ss.User)
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

	conf.Password = ss.CreatePasswd()
	go ss.RunNew(conf)
	err = <-ss.ConnChan
	if err != nil {
		Sandstorm.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Sandstorm.HTTPSuccess(w, conf.Port+":"+conf.Password)
}
