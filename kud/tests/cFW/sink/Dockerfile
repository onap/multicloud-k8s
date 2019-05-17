FROM ubuntu:16.04
MAINTAINER Victor Morales <electrocucaracha@gmail.com>

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

ENV demo_artifacts_version "1.5.0"

RUN apt-get update && apt-get install -y -qq wget net-tools

WORKDIR /opt

RUN wget "https://git.onap.org/demo/plain/vnfs/vFW/scripts/v_sink_init.sh" \
 && chmod +x v_sink_init.sh 

# NOTE: this script executes $ route add -net 192.168.10.0 netmask 255.255.255.0 gw 192.168.20.100
# which results in this error if doesn't have all nics required  -> SIOCADDRT: File exists
CMD ["./v_sink_init.sh"]
