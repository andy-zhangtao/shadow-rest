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

	email := `<h1>创建成功,用户信息如下:</h1> <h3> <br/> 加密方式:aes-256-cfb <br/><br/> ` + conf.Port + `:` + conf.Password + ` <br/><br/> 流量限制:` + strconv.Itoa(conf.Rate) + `KB <br/><br/> 过期时间:` + conf.Expriy + `D <br/><br/> <h2/> `
	err = ss.SendEmail(email, conf.Email)
	if err != nil {
		Sandstorm.HTTPSuccess(w, conf.Port+":"+conf.Password+" "+err.Error())
		return
	}
	Sandstorm.HTTPSuccess(w, conf.Port+":"+conf.Password)
}
