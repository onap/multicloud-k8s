# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

FROM ubuntu:16.04

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY
ENV no_proxy $NO_PROXY

ENV CSAR_DIR "/opt/csar"
ENV KUBE_CONFIG_DIR "/opt/kubeconfig"
ENV DATABASE_TYPE "consul"
ENV DATABASE_IP "127.0.0.1"
ENV OVN_CENTRAL_ADDRESS "127.0.0.1:6641"

EXPOSE 8081

RUN groupadd -r onap && useradd -r -g onap onap
RUN apt-get update && apt-get install -y -qq apt-transport-https curl \
 && echo "deb https://packages.wand.net.nz xenial main" > /etc/apt/sources.list.d/wand.list \
 && curl https://packages.wand.net.nz/keyring.gpg -o /etc/apt/trusted.gpg.d/wand.gpg \
 && apt-get update && apt install -y -qq ovn-common

WORKDIR /opt/multicloud/k8s
ADD ./k8plugin ./
ADD ./*.so ./
RUN chown onap:onap /opt/multicloud/k8s -R

USER onap

CMD ["./k8plugin"]
