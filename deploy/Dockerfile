FROM alpine:latest

RUN wget -O /shadowsocks-server "https://www.dropbox.com/s/irep7vw89zp7r87/shadowsocks-server?dl=0" && 
    chmod 755 /shadowsocks-server

CMD ["/shadowsocks-server"]