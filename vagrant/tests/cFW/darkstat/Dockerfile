FROM ubuntu:16.04
MAINTAINER Victor Morales <electrocucaracha@gmail.com>

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y -qq darkstat

EXPOSE 667

CMD ["/usr/sbin/darkstat", "-i", "eth1", "--no-daemon"]
