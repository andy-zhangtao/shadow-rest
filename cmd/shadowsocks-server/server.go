package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	ss "github.com/andy-zhangtao/shadow-rest/shadowsocks"
	"github.com/andy-zhangtao/shadow-rest/shadowsocks/handler"
)

func waitSignal(configFile string, config *ss.Config) {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	for sig := range sigChan {
		if sig == syscall.SIGHUP {
			ss.UpdatePasswd(configFile, config)
		} else {
			// is this going to happen?
			log.Printf("caught signal %v, exit", sig)
			os.Exit(0)
		}
	}
}

var configFile string
var config *ss.Config
var backConfig *ss.Config

func main() {
	// log.SetOutput(os.Stdout)

	// debug := ss.GetDebug()
	var cmdConfig ss.Config
	var printVer bool
	var core int

	flag.BoolVar(&printVer, "version", false, "print version")
	flag.StringVar(&configFile, "c", "config.json", "specify config file")
	flag.StringVar(&cmdConfig.Password, "k", "", "password")
	flag.IntVar(&cmdConfig.ServerPort, "p", 0, "server port")
	flag.IntVar(&cmdConfig.Timeout, "t", 300, "timeout in seconds")
	flag.StringVar(&cmdConfig.Method, "m", "", "encryption method, default: aes-256-cfb")
	flag.IntVar(&core, "core", 0, "maximum number of CPU cores to use, default is determinied by Go runtime")
	// flag.BoolVar((*bool)(&debug), "d", false, "print debug message")

	flag.Parse()

	if printVer {
		ss.PrintVersion()
		os.Exit(0)
	}

	// ss.SetDebug(debug)
	if strings.HasSuffix(cmdConfig.Method, "-auth") {
		cmdConfig.Method = cmdConfig.Method[:len(cmdConfig.Method)-5]
		cmdConfig.Auth = true
	}

	var err error
	config, err = ss.ParseConfig(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", configFile, err)
			os.Exit(1)
		}
		config = &cmdConfig
	} else {
		ss.UpdateConfig(config, &cmdConfig)
	}

	sa := config.GetServerArray()
	if len(sa) > 1 {
		ss.GlobaIP = sa[0]
	}

	if config.Method == "" {
		config.Method = "aes-256-cfb"
	}
	if err = ss.CheckCipherMethod(config.Method); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = ss.UnifyPortPassword(config); err != nil {
		os.Exit(1)
	}
	if core > 0 {
		runtime.GOMAXPROCS(core)
	}

	ss.Minport = config.Minport
	ss.Maxport = config.Maxport

	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	err = ss.ParseBackConfig(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	for port, password := range config.PortPassword {
		go ss.Run(port, password, config.Method, config.Auth)
	}

	// 统计端口链接信息
	go ss.HandleListen()
	// 统计端口数据流量
	go ss.HandleRate()
	// 网络链接失效期判断
	go ss.IsExpiry()
	// 网络链接流量判断
	go ss.IsAboveRate()
	// 用户数据持久化
	go ss.Persistence()
	// go ss.PersistencePasswd()

	// 重新加载配置文件
	go waitSignal(configFile, config)

	r := mux.NewRouter()

	r.HandleFunc("/user/all", handler.GetListenHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/stop/{ports}", handler.DeleteListenHandler).Methods(http.MethodDelete)
	r.HandleFunc("/user/restart", handler.RestartListenHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/expiry", handler.SetExpiryHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/info", handler.GetInfoHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/rate", handler.GetRateHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/rate", handler.SetRateHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/new", handler.CreateUserHandler).Methods(http.MethodPost)

	r.HandleFunc("/version", handler.GetVersion).Methods(http.MethodGet)

	r.HandleFunc("/proxy", handler.ProxyConnHandler).Methods(http.MethodGet)
	r.HandleFunc("/proxy/{id}", handler.ProxyInfo).Methods(http.MethodGet)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	log.Println(http.ListenAndServe(":8000", handlers.CORS(headersOk, originsOk, methodsOk)(r)))
}
