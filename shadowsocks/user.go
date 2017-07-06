package shadowsocks

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/andy-zhangtao/shadow-rest/shadowsocks/db"
	"github.com/andy-zhangtao/shadow-rest/shadowsocks/util"

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
	Port      string
	Expriy    string `json:"expriy"`
	Rate      int    `json:"rate"`
	RateLimit int    `json:"ratelimit"`
	Email     string `json:"email"` //16-12-14 新增email属性
	Password  string `json:"password"`
	ID        string `json:"port"`
}

// UserPass 将端口和链接口令单独保存
type UserPass struct {
	Port     string `json:"port"`
	Password string `json:"password"`
}

// UserPassBack 用于格式化输出UserPass
type UserPassBack struct {
	Upb []UserPass `json:"upb"`
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
	if u.ID == "" {
		u.Port = getNextPort()
	} else {
		u.Port = u.ID
	}
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
	if mongoSession == nil {
		mongoSession = db.GetMongo()
	}

	u := mongoSession.DB(os.Getenv(util.MONGODB)).C("user")
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case user := <-UserChan:
			log.Printf("%s\n", user.Port)
			err := u.Insert(&User{
				RateLimit: user.Rate,
				Password:  user.Password,
				Email:     user.Email,
				Expriy:    user.Expriy,
				ID:        user.Port,
			})
			if err != nil {
				log.Println(err.Error())
			}
		case <-ticker.C:
			for l := range listenMap {
				old := bson.M{"id": listenMap[l].Port}
				newListen := bson.M{"$set": bson.M{"rate": listenMap[l].Rate}}
				err := u.Update(old, newListen)
				if err != nil {
					log.Println(err.Error())
				}
			}
		case pp := <-PasswdChan:
			if len(PassMap) == 0 {
				PassMap = make(map[string]UserPass)
			}

			if pp.Password == "" {
				err := u.Remove(bson.M{"id": pp.Port})
				if err != nil {
					log.Println("Remove User Error:" + err.Error())
				}
			} else {
				PassMap[pp.Port] = *pp
			}
		}
	}
	// config := os.Getenv("configdir")
	// if config == "" {
	// 	config = "/config"
	// }

	// ticker := time.NewTicker(1 * time.Minute)
	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		lm := make([]Listen, len(listenMap))
	// 		i := 0
	// 		for l := range listenMap {
	// 			lm[i] = listenMap[l]
	// 			i++
	// 		}

	// 		lb := &ListenBak{
	// 			Lb: lm,
	// 		}

	// 		data, err := json.Marshal(lb)
	// 		if err != nil {
	// 			golog.Error(err.Error())
	// 		}

	// 		err = ioutil.WriteFile(config+"/user.json", data, 0600)
	// 		if err != nil {
	// 			golog.Error(err.Error())
	// 		}
	// 	}
	// }
}

// KillUserPass 当断开链接时，一并删除其口令
func KillUserPass(port string) {
	PasswdChan <- &UserPass{
		Port: port,
	}
}

// PersistencePasswd 用户口令持久化
func PersistencePasswd() {
	config := os.Getenv("configdir")
	if config == "" {
		config = "/config"
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

			up := make([]UserPass, len(PassMap))
			i := 0
			for p := range PassMap {
				up[i] = PassMap[p]
				i++
			}

			upb := &UserPassBack{
				Upb: up,
			}

			data, err := json.Marshal(upb)
			if err != nil {
				golog.Error(err.Error())
			}

			err = ioutil.WriteFile(config+"/passwd.json", data, 0600)
			if err != nil {
				golog.Error(err.Error())
			}
		}
	}
}
