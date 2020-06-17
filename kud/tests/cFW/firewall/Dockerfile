FROM ubuntu:18.04 as builder
MAINTAINER Victor Morales <electrocucaracha@gmail.com>

ENV demo_artifacts_version "1.6.0"
ENV repo_url "https://nexus.onap.org/content/repositories/staging/org/onap/demo/vnf"

RUN apt-get update && apt-get install -y -qq --no-install-recommends \
 wget ca-certificates

WORKDIR /opt

RUN wget "${repo_url}/sample-distribution/${demo_artifacts_version}/sample-distribution-${demo_artifacts_version}-hc.tar.gz" \
 && tar -zmxf sample-distribution-${demo_artifacts_version}-hc.tar.gz \
 && rm sample-distribution-${demo_artifacts_version}-hc.tar.gz \
 && mv sample-distribution-${demo_artifacts_version} honeycomb \
 && sed -i 's/"restconf-binding-address": .*/"restconf-binding-address": "0.0.0.0",/g' /opt/honeycomb/config/restconf.json

RUN apt-get install -y -qq --no-install-recommends \
 make gcc libc6-dev libcurl4-gnutls-dev

RUN wget "${repo_url}/ves5/ves/${demo_artifacts_version}/ves-${demo_artifacts_version}-demo.tar.gz" \
 && tar -zmxf ves-${demo_artifacts_version}-demo.tar.gz \
 && rm ves-${demo_artifacts_version}-demo.tar.gz \
 && mv ves-${demo_artifacts_version} VES

RUN wget "${repo_url}/ves5/ves_vfw_reporting/${demo_artifacts_version}/ves_vfw_reporting-${demo_artifacts_version}-demo.tar.gz" \
 && tar -zmxf ves_vfw_reporting-${demo_artifacts_version}-demo.tar.gz \
 && rm ves_vfw_reporting-${demo_artifacts_version}-demo.tar.gz \
 && mv ves_vfw_reporting-${demo_artifacts_version} VES/evel/evel-library/code/VESreporting \
 && chmod +x VES/evel/evel-library/code/VESreporting/go-client.sh \
 && make -C /opt/VES/evel/evel-library/bldjobs/

FROM vpp

COPY --from=builder /opt/honeycomb /opt/honeycomb
COPY --from=builder /opt/VES/evel/evel-library/code/VESreporting /opt/VESreporting
COPY --from=builder /opt/VES/evel/evel-library/libs/x86_64/libevel.so /usr/lib/x86_64-linux-gnu/
COPY init.sh /opt/init.sh

ENV DCAE_COLLECTOR_IP ""
ENV DCAE_COLLECTOR_PORT ""

RUN apt-get update && apt-get install -y -qq --no-install-recommends \
 openjdk-8-jre iproute2 libcurl4-gnutls-dev

ENTRYPOINT ["/bin/bash"]
CMD ["/opt/init.sh"]
