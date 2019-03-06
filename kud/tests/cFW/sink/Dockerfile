FROM ubuntu:16.04
MAINTAINER Victor Morales <electrocucaracha@gmail.com>

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

ENV protected_net_cidr "192.168.20.0/24"
ENV fw_ipaddr "192.168.10.100"
ENV sink_ipaddr "192.168.20.250"
ENV demo_artifacts_version "1.3.0"
ENV protected_net_gw "192.168.20.100"
ENV unprotected_net "192.168.10.0/24"

RUN apt-get update && apt-get install -y -qq wget net-tools

WORKDIR /opt

RUN wget "https://git.onap.org/demo/plain/vnfs/vFW/scripts/v_sink_init.sh" \
 && chmod +x v_sink_init.sh 

RUN mkdir -p config/ \
 && echo $protected_net_cidr > config/protected_net_cidr.txt \
 && echo $fw_ipaddr > config/fw_ipaddr.txt \
 && echo $sink_ipaddr > config/sink_ipaddr.txt \
 && echo $demo_artifacts_version > config/demo_artifacts_version.txt \
 && echo $protected_net_gw > config/protected_net_gw.txt \
 && echo $unprotected_net > config/unprotected_net.txt

# NOTE: this script executes $ route add -net 192.168.10.0 netmask 255.255.255.0 gw 192.168.20.100
# which results in this error if doesn't have all nics required  -> SIOCADDRT: File exists
CMD ["./v_sink_init.sh"]
