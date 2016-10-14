package shadowsocks

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
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
	for index := 0; index < 7; index++ {
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
