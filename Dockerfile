FROM centos:7

ADD bin/shadowsocks-server-linux64-1.1.5.gz /shadowsocks-server-linux64-1.1.5.gz
RUN gzip -df /shadowsocks-server-linux64-1.1.5.gz ; chmod 755 /shadowsocks-server-linux64-1.1.5

CMD ["/shadowsocks-server-linux64-1.1.5"]
