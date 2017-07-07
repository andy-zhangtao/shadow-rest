package shadowsocks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/andy-zhangtao/shadow-rest/shadowsocks/db"

	"github.com/andy-zhangtao/shadow-rest/shadowsocks/util"

	mgo "gopkg.in/mgo.v2"

	// "log"
	"os"
	"reflect"
	"strings"
	"time"
)

// Config 网络链接配置信息
type Config struct {
	Server     interface{} `json:"server"`
	ServerPort int         `json:"server_port"`
	LocalPort  int         `json:"local_port"`
	Password   string      `json:"password"`
	Method     string      `json:"method"` // encryption method
	Auth       bool        `json:"auth"`   // one time auth
	Minport    int         `json:"minport"`
	Maxport    int         `json:"maxport"`
	// following options are only used by server
	PortPassword map[string]string `json:"port_password"`
	Timeout      int               `json:"timeout"`

	// following options are only used by client

	// The order of servers in the client config is significant, so use array
	// instead of map to preserve the order.
	ServerPassword [][]string `json:"server_password"`
}

var readTimeout time.Duration
var listenBakConf map[string]Listen

var mongoSession *mgo.Session

// GetServerArray 获取当前所有服务参数
func (config *Config) GetServerArray() []string {
	// Specifying multiple servers in the "server" options is deprecated.
	// But for backward compatiblity, keep this.
	if config.Server == nil {
		return nil
	}
	single, ok := config.Server.(string)
	if ok {
		return []string{single}
	}
	arr, ok := config.Server.([]interface{})
	if ok {
		/*
			if len(arr) > 1 {
				log.Println("Multiple servers in \"server\" option is deprecated. " +
					"Please use \"server_password\" instead.")
			}
		*/
		serverArr := make([]string, len(arr), len(arr))
		for i, s := range arr {
			serverArr[i], ok = s.(string)
			if !ok {
				goto typeError
			}
		}
		return serverArr
	}
typeError:
	panic(fmt.Sprintf("Config.Server type error %v", reflect.TypeOf(config.Server)))
}

// ParseConfig 解析配置文件数据
func ParseConfig(path string) (config *Config, err error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	config = &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	readTimeout = time.Duration(config.Timeout) * time.Second
	if strings.HasSuffix(strings.ToLower(config.Method), "-auth") {
		config.Method = config.Method[:len(config.Method)-5]
		config.Auth = true
	}
	return
}

// SetDebug 设置输出日志级别
func SetDebug(d DebugLog) {
	Debug = d
}

//  UpdateConfig  Useful for command line to override options specified in config file Debug is not updated.
func UpdateConfig(old, new *Config) {
	// Using reflection here is not necessary, but it's a good exercise.
	// For more information on reflections in Go, read "The Laws of Reflection"
	// http://golang.org/doc/articles/laws_of_reflection.html
	newVal := reflect.ValueOf(new).Elem()
	oldVal := reflect.ValueOf(old).Elem()

	// typeOfT := newVal.Type()
	for i := 0; i < newVal.NumField(); i++ {
		newField := newVal.Field(i)
		oldField := oldVal.Field(i)
		// log.Printf("%d: %s %s = %v\n", i,
		// typeOfT.Field(i).Name, newField.Type(), newField.Interface())
		switch newField.Kind() {
		case reflect.Interface:
			if fmt.Sprintf("%v", newField.Interface()) != "" {
				oldField.Set(newField)
			}
		case reflect.String:
			s := newField.String()
			if s != "" {
				oldField.SetString(s)
			}
		case reflect.Int:
			i := newField.Int()
			if i != 0 {
				oldField.SetInt(i)
			}
		}
	}

	old.Timeout = new.Timeout
	readTimeout = time.Duration(old.Timeout) * time.Second
}

// ParseBackConfig 解析备份配置文件数据
func ParseBackConfig(config *Config) error {
	if mongoSession == nil {
		mongoSession = db.GetMongo()
	}

	var user []User
	u := mongoSession.DB(os.Getenv(util.MONGODB)).C("user")
	err := u.Find(nil).All(&user)
	if err != nil {
		return err
	}

	pb := make(map[string]string)
	listenBakConf = make(map[string]Listen)

	for _, us := range user {
		log.Println(us)
		pb[us.ID] = us.Password
		listenBakConf[us.ID] = Listen{
			Port:       us.ID,
			Rate:       us.Rate,
			ExpiryDate: us.Expriy,
			RateLimit:  us.RateLimit,
			Email:      us.Email,
		}
	}

	config.PortPassword = pb

	// con := os.Getenv("configdir")
	// if con == "" {
	// 	con = "/config"
	// }

	// // 解析口令配置文件
	// file, err := os.Open(con + "/passwd.json") // For read access.
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// data, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	return err
	// }

	// up := &UserPassBack{}
	// if err = json.Unmarshal(data, up); err != nil {
	// 	return err
	// }

	// pb := make(map[string]string)
	// for _, b := range up.Upb {
	// 	pb[b.Port] = b.Password
	// }

	// config.PortPassword = pb

	// // 解析网络备份文件
	// file, err = os.Open(con + "/user.json") // For read access.
	// if err != nil {
	// 	return err
	// }

	// data, err = ioutil.ReadAll(file)
	// if err != nil {
	// 	return err
	// }

	// lb := &ListenBak{}
	// if err = json.Unmarshal(data, lb); err != nil {
	// 	return err
	// }

	// listenBakConf = make(map[string]Listen)
	// for _, l := range lb.Lb {
	// 	listenBakConf[l.Port] = l
	// }
	return nil
}
