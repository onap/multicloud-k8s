FROM openwrt-1806-4-base

#EXPOSE 80
ENV http_proxy={docker_proxy}
ENV https_proxy={docker_proxy}
ENV no_proxy=localhost,120.0.0.1,192.168.*

RUN mkdir /var/lock && \
    opkg update && \
    opkg install uhttpd-mod-lua && \
    uci set uhttpd.main.interpreter='.lua=/usr/bin/lua' && \
    uci commit uhttpd && \
    opkg install mwan3 && \
    opkg install luci-app-mwan3; exit 0

COPY system /etc/config/system
COPY commands.lua /usr/lib/lua/luci/controller/

ENV http_proxy=
ENV https_proxy=
ENV no_proxy=

USER root

# using exec format so that /sbin/init is proc 1 (see procd docs)
CMD ["/sbin/init"]
