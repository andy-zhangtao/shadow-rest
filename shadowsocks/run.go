package shadowsocks

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/andy-zhangtao/shadow-rest/configure"
)

var debug DebugLog

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip addres start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4 = 1 // type is ipv4 address
	typeDm   = 3 // type is domain address
	typeIPv6 = 4 // type is ipv6 address

	lenIPv4     = net.IPv4len + 2 // ipv4 + 2port
	lenIPv6     = net.IPv6len + 2 // ipv6 + 2port
	lenDmBase   = 2               // 1addrLen + 2port, plus addrLen
	lenHmacSha1 = 10
	logCntDelta = 100
)

var connCnt int
var nextLogConnCnt = logCntDelta

// listenChan 用于统计网络链接信息
var listenChan = make(chan *Listen)

// rateChan 用于统计网络链接流量信息
var rateChan = make(chan *Listen)

// ConnChan 用于定制网络链接信息
var ConnChan = make(chan error)

// PasswdChan 用于保存网络链接及其口令
var PasswdChan = make(chan *UserPass)

// UserChan 传递新建用户信息
var UserChan = make(chan *User)

var passwdManager = PasswdManager{PortListener: map[string]*PortListener{}}

// GetDebug 用于设置Debug函数
func GetDebug() DebugLog {
	return debug
}

func getRequest(conn *Conn, auth bool) (host string, ota bool, err error) {
	SetReadTimeout(conn)

	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(addrType) + 1(lenByte) + 256(max length address) + 2(port) + 10(hmac-sha1)
	buf := make([]byte, 270)
	// read till we get possible domain length field
	if _, err = io.ReadFull(conn, buf[:idType+1]); err != nil {
		return
	}

	var reqStart, reqEnd int
	addrType := buf[idType]
	switch addrType & AddrMask {
	case typeIPv4:
		reqStart, reqEnd = idIP0, idIP0+lenIPv4
	case typeIPv6:
		reqStart, reqEnd = idIP0, idIP0+lenIPv6
	case typeDm:
		if _, err = io.ReadFull(conn, buf[idType+1:idDmLen+1]); err != nil {
			return
		}
		reqStart, reqEnd = idDm0, int(idDm0+buf[idDmLen]+lenDmBase)
	default:
		err = fmt.Errorf("addr type %d not supported", addrType&AddrMask)
		return
	}

	if _, err = io.ReadFull(conn, buf[reqStart:reqEnd]); err != nil {
		return
	}

	// Return string for typeIP is not most efficient, but browsers (Chrome,
	// Safari, Firefox) all seems using typeDm exclusively. So this is not a
	// big problem.
	switch addrType & AddrMask {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}
	// parse port
	port := binary.BigEndian.Uint16(buf[reqEnd-2 : reqEnd])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	// if specified one time auth enabled, we should verify this
	if auth || addrType&OneTimeAuthMask > 0 {
		ota = true
		if _, err = io.ReadFull(conn, buf[reqEnd:reqEnd+lenHmacSha1]); err != nil {
			return
		}
		iv := conn.GetIv()
		key := conn.GetKey()
		actualHmacSha1Buf := HmacSha1(append(iv, key...), buf[:reqEnd])
		if !bytes.Equal(buf[reqEnd:reqEnd+lenHmacSha1], actualHmacSha1Buf) {
			err = fmt.Errorf("verify one time auth failed, iv=%v key=%v data=%v", iv, key, buf[:reqEnd])
			return
		}
	}
	return
}

// HandleConnection 处理具体链接数据
func HandleConnection(conn *Conn, auth bool, lr *Listen) {
	var host string

	connCnt++ // this maybe not accurate, but should be enough
	if connCnt-nextLogConnCnt >= 0 {
		// XXX There's no xadd in the atomic package, so it's difficult to log
		// the message only once with low cost. Also note nextLogConnCnt maybe
		// added twice for current peak connection number level.
		log.Printf("Number of client connections reaches %d\n", nextLogConnCnt)
		nextLogConnCnt += logCntDelta
	}

	// function arguments are always evaluated, so surround debug statement
	// with if statement
	if debug {
		debug.Printf("new client %s->%s\n", conn.RemoteAddr().String(), conn.LocalAddr())
	}

	closed := false
	defer func() {
		if debug {
			debug.Printf("closed pipe %s<->%s\n", conn.RemoteAddr(), host)
		}
		connCnt--
		if !closed {
			conn.Close()
		}
	}()

	host, ota, err := getRequest(conn, auth)
	if err != nil {
		log.Println("error getting request", conn.RemoteAddr(), conn.LocalAddr(), err)
		return
	}

	debug.Println("connecting", host)
	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			log.Println("dial error:", err)
		} else {
			log.Println("error connecting to:", host, err)
		}
		return
	}

	defer func() {
		if !closed {
			remote.Close()
		}
	}()

	if debug {
		debug.Printf("piping %s<->%s ota=%v connOta=%v", conn.RemoteAddr(), host, ota, conn.IsOta())
	}
	if ota {
		go PipeThenCloseOta(conn, remote, lr)
	} else {
		go PipeThenClose(conn, remote, lr)
	}

	PipeThenClose(remote, conn, lr)
	closed = true
	return
}

// RunNew 通过API自动创建一个网络链接
func RunNew(u *User) {
	CreateUser(u)

	var ln net.Listener
	var err error
	if GlobaIP != "" {
		ln, err = net.Listen("tcp6", GlobaIP+":"+u.Port)
	} else {
		ln, err = net.Listen("tcp4", ":"+u.Port)
	}

	if err != nil {
		log.Printf("error listening port %v: %v\n", u.Port, err)
		ConnChan <- err
		return
	}

	passwdManager.Add(u.Port, u.Password, ln)
	var cipher *Cipher
	log.Printf("server listening port %v ...\n", u.Port)

	d, err := strconv.Atoi(u.Expriy)
	if err != nil {
		log.Printf("error expriy date %v: %v\n", u.Expriy, err)
		ConnChan <- err
		return
	}

	ed := time.Now().AddDate(0, 0, d).Format(TIMEFORMATE)
	ls := &Listen{
		Port:       u.Port,
		listen:     ln,
		ExpiryDate: ed,
		RateLimit:  u.Rate,
		Email:      u.Email,
	}

	listenChan <- ls

	lr := &Listen{
		Port: u.Port,
		Rate: 0,
	}

	// 更新用户失效日期
	u.Expriy = ls.ExpiryDate
	ConnChan <- nil
	UserChan <- u
	// PasswdChan <- &UserPass{
	// 	Port:     u.Port,
	// 	Password: u.Password,
	// }

	for {
		conn, err := ln.Accept()
		if err != nil {
			// listener maybe closed to update password
			debug.Printf("accept error: %v\n", err)
			return
		}

		// Creating cipher upon first connection.
		if cipher == nil {
			log.Println("creating cipher for port:", u.Port)
			cipher, err = NewCipher(configure.DEFAULTMETHOD, u.Password)
			if err != nil {
				log.Printf("Error generating cipher for port: %s %v\n", u.Port, err)
				conn.Close()
				continue
			}
		}

		go HandleConnection(NewConn(conn, cipher.Copy()), false, lr)
	}
}

// Run SS执行函数
// @port    监听端口
// @passwd  端口绑定密钥
// @method  端口绑定加密算法
// @auth    是否需要一次性认证
func Run(port, password, method string, auth bool) {

	var ln net.Listener
	var err error
	if GlobaIP != "" {
		ln, err = net.Listen("tcp6", GlobaIP+":"+port)
	} else {
		ln, err = net.Listen("tcp4", ":"+port)
	}

	if err != nil {
		log.Printf("error listening port %v: %v\n", port, err)
		return
	}

	passwdManager.Add(port, password, ln)
	var cipher *Cipher
	log.Printf("server listening port %v ...\n", port)

	ed := ""
	rate := 0
	rateLimit := 0
	if listenBakConf[port].Port != "" {
		ed = listenBakConf[port].ExpiryDate
		rate = listenBakConf[port].Rate
		rateLimit = listenBakConf[port].RateLimit
	} else {
		ed = time.Now().AddDate(0, 0, 7).Format(TIMEFORMATE)
	}

	ls := &Listen{
		Port:       port,
		listen:     ln,
		ExpiryDate: ed,
		Rate:       rate,
		RateLimit:  rateLimit,
	}
	listenChan <- ls

	lr := &Listen{
		Port: port,
	}

	// PasswdChan <- &UserPass{
	// 	Port:     port,
	// 	Password: password,
	// }

	for {
		conn, err := ln.Accept()
		if err != nil {
			// listener maybe closed to update password
			debug.Printf("accept error: %v\n", err)
			return
		}

		// Creating cipher upon first connection.
		if cipher == nil {
			log.Println("creating cipher for port:", port)
			cipher, err = NewCipher(method, password)
			if err != nil {
				log.Printf("Error generating cipher for port: %s %v\n", port, err)
				conn.Close()
				continue
			}
		}

		go HandleConnection(NewConn(conn, cipher.Copy()), auth, lr)
	}
}

// UpdatePasswd 更新链接密钥
func UpdatePasswd(configFile string, config *Config) {
	log.Println("updating password")
	newconfig, err := ParseConfig(configFile)
	if err != nil {
		log.Printf("error parsing config file %s to update password: %v\n", configFile, err)
		return
	}
	oldconfig := config
	config = newconfig

	if err = UnifyPortPassword(config); err != nil {
		return
	}
	for port, passwd := range config.PortPassword {
		passwdManager.UpdatePortPasswd(port, passwd, config.Auth)
		if oldconfig.PortPassword != nil {
			delete(oldconfig.PortPassword, port)
		}
	}
	// port password still left in the old config should be closed
	for port := range oldconfig.PortPassword {
		log.Printf("closing port %s as it's deleted\n", port)
		passwdManager.Del(port)
	}
	log.Println("password updated")
}

// UnifyPortPassword 确认链接密钥
func UnifyPortPassword(config *Config) (err error) {
	if len(config.PortPassword) == 0 { // this handles both nil PortPassword and empty one
		if !EnoughOptions(config) {
			fmt.Fprintln(os.Stderr, "must specify both port and password")
			return errors.New("not enough options")
		}
		port := strconv.Itoa(config.ServerPort)
		config.PortPassword = map[string]string{port: config.Password}
	} else {
		if config.Password != "" || config.ServerPort != 0 {
			fmt.Fprintln(os.Stderr, "given port_password, ignore server_port and password option")
		}
	}
	return
}

// EnoughOptions 确认参数是否完整
func EnoughOptions(config *Config) bool {
	return config.ServerPort != 0 && config.Password != ""
}

// HandleListen 控制每个网络链接
func HandleListen() {
	for {
		select {
		case ls := <-listenChan:
			AddListen(*ls)
		}
	}
}

// HandleRate 统计每个端口的流出流量
func HandleRate() {
	for {
		select {
		case lr := <-rateChan:
			AddRate(*lr)
		}
	}
}
