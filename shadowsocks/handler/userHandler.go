package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

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

	if conf.Email == "" {
		conf.Email = os.Getenv("SS_DEST_EMAIL") //如果没有指定用户Email，使用管理员邮箱进行替代
	}

	conf.Password = ss.CreatePasswd()
	go ss.RunNew(conf)
	err = <-ss.ConnChan
	if err != nil {
		Sandstorm.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := `<h1>Your account info is belowing:</h1> <h3> <br/> encrypt:aes-256-cfb <br/><br/> ` + conf.Port + `:` + conf.Password + ` <br/><br/> Rate:` + strconv.Itoa(conf.Rate) + `KB <br/><br/> Expriy:` + conf.Expriy + `D <br/><br/> <h2/> `
	err = ss.SendEmail(email, conf.Email)
	if err != nil {
		Sandstorm.HTTPSuccess(w, conf.Port+":"+conf.Password+" "+err.Error())
		return
	}
	Sandstorm.HTTPSuccess(w, conf.Port+":"+conf.Password)
}
