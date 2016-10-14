package shadowsocks

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/andy-zhangtao/golog"
)

/**
 * Created with VScode.
 * User: andy.zhangtao <ztao8607@gmail.com>
 * Date: 16-10-14
 * Time: 11:45
 * 创建新链接
 * PS: 一个账户对应一个网络链接，但一个链接不限制终端数量
 */

// User 网络链接结构体. Conn对应网络链接，所以这里使用User替代Conn
type User struct {
	Port     string
	Expriy   string `json:"expriy"`
	Rate     int    `json:"rate"`
	Password string
}

// UserPass 将端口和链接口令单独保存
type UserPass struct {
	Port     string `json:"port"`
	Password string `json:"password"`
}

var (
	// Minport 默认最低端口从10001开始
	Minport = 10001
	// Maxport 默认最低端口从19999开始
	Maxport = 19999
	// Currport 当前端口
	Currport = Minport
)

var mutex sync.Mutex

// CreateUser 创建新用户
func CreateUser(u *User) {
	u.Port = getNextPort()
}

// CreatePasswd 创建一个8位数随机密码
func CreatePasswd() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	passwd := ""
	for {
		if len(passwd) == 8 {
			break
		}
		p := strconv.Itoa(r.Intn(10))
		passwd = passwd + p
	}

	return passwd
}

// getNextPort 获取下一个有效端口
func getNextPort() string {
	mutex.Lock()

	defer func() {
		Currport++
		mutex.Unlock()
	}()

	for {
		if Currport > Maxport {
			Currport = Minport
		}

		if listenMap[strconv.Itoa(Currport)].Port == "" {
			return strconv.Itoa(Currport)
		}
		Currport++
	}
}

// Persistence 用户数据持久化
func Persistence() {
	usr, err := user.Current()
	if err != nil {
		golog.Error(err.Error())
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			data, err := json.Marshal(listenMap)
			if err != nil {
				golog.Error(err.Error())
			}

			err = ioutil.WriteFile(usr.HomeDir+"/user.json", data, 0600)
			if err != nil {
				golog.Error(err.Error())
			}
		}
	}
}

// KillUserPass 当断开链接时，一并删除其口令
func KillUserPass(port string) {
	PasswdChan <- &UserPass{
		Port: port,
	}
}

// PersistencePasswd 用户口令持久化
func PersistencePasswd() {
	usr, err := user.Current()
	if err != nil {
		golog.Error(err.Error())
		return
	}

	for {
		select {
		case pp := <-PasswdChan:
			if len(PassMap) == 0 {
				PassMap = make(map[string]UserPass)
			}

			if pp.Password == "" {
				// 删除
				delete(PassMap, pp.Port)
			} else {
				PassMap[pp.Port] = *pp
			}

			data, err := json.Marshal(PassMap)
			if err != nil {
				golog.Error(err.Error())
			}

			err = ioutil.WriteFile(usr.HomeDir+"/passwd.json", data, 0600)
			if err != nil {
				golog.Error(err.Error())
			}
		}
	}
}
