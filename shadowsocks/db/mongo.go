package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andy-zhangtao/shadow-rest/shadowsocks/util"

	mgo "gopkg.in/mgo.v2"
)

var Session *mgo.Session

func init() {
	var err error

	if os.Getenv(util.MONGOURL) == "" {
		return
	}

	if Session == nil {
		login := &mgo.DialInfo{
			Addrs:    []string{os.Getenv(util.MONGOURL)},
			Timeout:  3600 * time.Second,
			Database: os.Getenv(util.MONGODB),
			Username: os.Getenv(util.USERNAME),
			Password: os.Getenv(util.PASSWORD),
		}

		// log.Printf("Connectting mongodb,env: [%v] host:[%s] db:[%s] \n", login.Addrs, login.Database, os.Environ())
		Session, err = mgo.DialWithInfo(login)
		if err != nil {
			fmt.Println(err.Error())
		}

		if err := Session.Ping(); err == nil {
			log.Printf("MONGO CONNECT SUCCESS! [%s]\n", Session.LiveServers())
		} else {
			log.Printf("MONGO CONNECT FAILED!! [%s] Error Info [%s]\n", os.Getenv(util.MONGOURL), err.Error())
		}

	}
}

func GetMongo() *mgo.Session {

	Session.Refresh()
	session := Session.Clone()
	if session == nil {
		fmt.Println("MONGODB SESSION IS NIL!!")
		return nil
	}

	return session
}
