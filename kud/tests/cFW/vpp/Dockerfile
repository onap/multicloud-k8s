FROM ubuntu:18.04
MAINTAINER Victor Morales <electrocucaracha@gmail.com>

ENV VERSION "19.01.2-release"

RUN apt-get update \
 && apt-get install -y -qq --no-install-recommends curl ca-certificates gnupg2 \
 && echo "deb [trusted=yes] https://packagecloud.io/fdio/release/ubuntu bionic main" | tee /etc/apt/sources.list.d/99fd.io.list \
 && curl -L https://packagecloud.io/fdio/release/gpgkey | apt-key add - \
 && mkdir -p /var/log/vpp/ \
 && apt-get update \
 && apt-get install -y -qq --no-install-recommends vpp=$VERSION vpp-lib=$VERSION vpp-plugins=$VERSION

COPY startup.conf /etc/vpp/startup.conf

CMD ["/usr/bin/vpp", "-c", "/etc/vpp/startup.conf"]
