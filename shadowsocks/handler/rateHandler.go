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
	"github.com/andy-zhangtao/shadow-rest/shadowsocks/util"
	"io/ioutil"
	"net/http"
	
	"github.com/andy-zhangtao/shadow-rest/configure"
	
	ss "github.com/andy-zhangtao/shadow-rest/shadowsocks"
)

// GetInfoHandler 获取当前所有的端口信息
func GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := ss.GetRate()
	if err != nil {
		util.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	content, _ := json.Marshal(keys)
	
	util.HTTPSuccess(w, string(content))
}

// GetRateHandler 获取当前所有端口流量数据
func GetRateHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := ss.GetRate()
	if err != nil {
		util.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	rate := make([]*ss.Rate, len(keys))
	
	for i, k := range keys {
		r := &ss.Rate{
			Port: k.ID,
			Rate: ss.ConvertRate(k.Rate),
		}
		rate[i] = r
	}
	
	content, _ := json.Marshal(rate)
	
	util.HTTPSuccess(w, string(content))
}

// SetRateHandler 设置网络最大流量
func SetRateHandler(w http.ResponseWriter, r *http.Request) {
	conf := new(ss.Rate)
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	err = json.Unmarshal(content, &conf)
	if err != nil {
		util.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if conf.Port == "" {
		util.HTTPError(w, configure.NOPORT, http.StatusInternalServerError)
		return
	}
	
	if conf.Rate == "" {
		util.HTTPError(w, configure.NORATE, http.StatusInternalServerError)
		return
	}
	
	err = ss.SetRate(conf.Port, conf.Rate)
	if err != nil {
		util.HTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	util.HTTPSuccess(w, "OK")
}
