# shadowsocks-go

Forked from shadowsocks/shadowsocks-go

shadowsocks-go version: 1.1.5  shadow-rest version: 0.1.X [![Build Status](https://travis-ci.org/andy-zhangtao/shadow-rest.svg?branch=master)]

## New Features

* Support REST API
* Support Rate count

### Rest API

* User API

| Endpoint | Usage |
| -------- | ----- |
| /user/all | Get all the proxy |
| /user/stop/{ports} | Stop specify proxy |
| /user/restart | Restart specify proxy |
| /user/expiry | Modify specify proxy expiry date |
| /user/rate | Get all the proxy rate |
| /user/info | Get all the proxy info |
| /user/new | Create a new user |

* Misc API

| Endpoint | Usage |
| -------- | ----- |
| /version | Get current version |

#####/user/new  POST
```
创建新账户
schema:
{
    "expriy":"有效期，string"，
    "rate":"数据量,单位字节。 int。 0 表示无限制"
    "email":"用户邮箱地址"
}

example:
{
        "expriy":"7",//7天有效期，以当天开始计算,
        "rate":0 //无限流量,
        "email":"ztao8607@gmail.com"
}
```

#####/user/expiry PUT
```
修改账户有效期
schema:
{
    "expriy":"有效期，string"，
    "port":"需要维护链接有效期的端口, string"
}

example
{
    "expiry":"2016-11-14",
    "port":"10001"
}
```

##### About Sand email.

If you wants send event email(rate/expriy), you should set the follow env :

| Name | Usage |
| -------- | ----- |
| SS_EMAIL_HOST | Host addr |
| SS_USER_NAME | Username |
| SS_PASS_WORD | Password |
| SS_PORT | Port (587 default)|
| SS_DEST_EMAIL | Dest Email addr |
| SS_ID | Host ID |



shadowsocks-go is a lightweight tunnel proxy which can help you get through firewalls. It is a port of [shadowsocks](https://github.com/clowwindy/shadowsocks).

The protocol is compatible with the origin shadowsocks (if both have been upgraded to the latest version).

**Note `server_password` option syntax changed in 0.6.2, the client now connects to servers in the order specified in the config.**

**Please develop on the latest develop branch if you want to send pull request.**

# Install

Download precompiled binarys from the [release page](https://github.com/shadowsocks/shadowsocks-go/releases). (All compiled with cgo disabled, except the mac version.)

You can also install from source (assume you have go installed):

```
# on server
go get github.com/shadowsocks/shadowsocks-go/cmd/shadowsocks-server
# on client
go get github.com/shadowsocks/shadowsocks-go/cmd/shadowsocks-local
```

It's recommended to disable cgo when compiling shadowsocks-go. This will prevent the go runtime from creating too many threads for dns lookup.

# Usage

Both the server and client program will look for `config.json` in the current directory. You can use `-c` option to specify another configuration file.

Configuration file is in json format and has the same syntax with [shadowsocks-nodejs](https://github.com/clowwindy/shadowsocks-nodejs/). You can download the sample [`config.json`](https://github.com/shadowsocks/shadowsocks-go/blob/master/config.json), change the following values:

```
server          your server ip or hostname
server_port     server port
local_port      local socks5 proxy port
method          encryption method, null by default (table), the following methods are supported:
                    aes-128-cfb, aes-192-cfb, aes-256-cfb, bf-cfb, cast5-cfb, des-cfb, rc4-md5, chacha20, salsa20, rc4, table
password        a password used to encrypt transfer
timeout         server option, in seconds
```

Run `shadowsocks-server` on your server. To run it in the background, run `shadowsocks-server > log &`.

On client, run `shadowsocks-local`. Change proxy settings of your browser to

```
SOCKS5 127.0.0.1:local_port
```

## About encryption methods

AES is recommended for shadowsocks-go. [Intel AES Instruction Set](http://en.wikipedia.org/wiki/AES_instruction_set) will be used if available and can make encryption/decryption very fast. To be more specific, **`aes-128-cfb` is recommended as it is faster and [secure enough](https://www.schneier.com/blog/archives/2009/07/another_new_aes.html)**.

**rc4 and table encryption methods are deprecated because they are not secure.**

### One Time Auth

Append `-auth` to the encryption method to enable [One Time Auth (OTA)](https://shadowsocks.org/en/spec/one-time-auth.html).

- For server: this will **force client use OTA**, non-OTA connection will be dropped. Otherwise, both OTA and non-OTA clients can connect
- For client: the `-A` command line option can also enable OTA

## Command line options

Command line options can override settings from configuration files. Use `-h` option to see all available options.

```
shadowsocks-local -s server_address -p server_port -k password
    -m aes-128-cfb -c config.json
    -b local_address -l local_port
shadowsocks-server -p server_port -k password
    -m aes-128-cfb -c config.json
    -t timeout
```

Use `-d` option to enable debug message.

## Use multiple servers on client

```
server_password    specify multiple server and password, server should be in the form of host:port
```

Here's a sample configuration [`client-multi-server.json`](https://github.com/shadowsocks/shadowsocks-go/blob/master/sample-config/client-multi-server.json). Given `server_password`, client program will ignore `server_port`, `server` and `password` options.

Servers are chosen in the order specified in the config. If a server can't be connected (connection failure), the client will try the next one. (Client will retry failed server with some probability to discover server recovery.)

## Multiple users with different passwords on server

The server can support users with different passwords. Each user will be served by a unique port. Use the following options on the server for such setup:

```
port_password   specify multiple ports and passwords to support multiple users
```

Here's a sample configuration [`server-multi-port.json`](https://github.com/shadowsocks/shadowsocks-go/blob/master/sample-config/server-multi-port.json). Given `port_password`, server program will ignore `server_port` and `password` options.

### Update port password for a running server

Edit the config file used to start the server, then send `SIGHUP` to the server process.

# Note to OpenVZ users

**Use OpenVZ VM that supports vswap**. Otherwise, the OS will incorrectly account much more memory than actually used. shadowsocks-go on OpenVZ VM with vswap takes about 3MB memory after startup. (Refer to [this issue](https://github.com/shadowsocks/shadowsocks-go/issues/3) for more details.)

If vswap is not an option and memory usage is a problem for you, try [shadowsocks-libev](https://github.com/madeye/shadowsocks-libev).
