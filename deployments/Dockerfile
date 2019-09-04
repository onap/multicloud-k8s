# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

FROM ubuntu:18.04

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY
ENV no_proxy $NO_PROXY

EXPOSE 9015

RUN groupadd -r onap && useradd -r -g onap onap

WORKDIR /opt/multicloud/k8splugin
RUN chown onap:onap /opt/multicloud/k8splugin -R

ADD --chown=onap ./k8plugin ./
ADD --chown=onap ./*.so ./plugins/

USER onap

CMD ["./k8plugin"]
