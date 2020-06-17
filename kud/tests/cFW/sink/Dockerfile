FROM ubuntu:18.04
MAINTAINER Ritu Sood <ritu.sood@intel.com>

COPY init.sh /opt/init.sh

ENV PROTECTED_NET_GW "192.168.20.100"
ENV UNPROTECTED_NET "192.168.10.0/24"

RUN apt-get update && apt-get install -y -qq --no-install-recommends \
 iproute2 darkstat
EXPOSE 667

ENTRYPOINT ["/bin/bash"]
CMD ["/opt/init.sh"]
