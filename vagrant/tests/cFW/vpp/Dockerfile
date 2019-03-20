FROM ubuntu:16.04
MAINTAINER Victor Morales <electrocucaracha@gmail.com>

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y -qq apt-transport-https \
 && echo "deb [trusted=yes] https://nexus.fd.io/content/repositories/fd.io.stable.1609.ubuntu.xenial.main/ ./" | tee -a /etc/apt/sources.list.d/99fd.io.list \
 && apt-get update \
 && apt-get install -y -qq vpp vpp-lib vpp-plugins 

COPY 80-vpp.conf /etc/sysctl.d/80-vpp.conf

CMD ["/usr/bin/vpp", "-c", "/etc/vpp/startup.conf"]
