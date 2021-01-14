FROM ubuntu:18.04 as base
ARG KUD_ENABLE_TESTS=false
ARG KUD_PLUGIN_ENABLED=false
ARG CONTAINER_RUNTIME=docker
ENV KUD_ENABLE_TESTS=$KUD_ENABLE_TESTS
ENV KUD_PLUGIN_ENABLED=$KUD_PLUGIN_ENABLED
ENV CONTAINER_RUNTIME=$CONTAINER_RUNTIME
ADD . /usr/src/multicloud-k8s
USER root
SHELL ["/bin/bash", "-c"]
WORKDIR /usr/src/multicloud-k8s/kud/hosting_providers/containerized
RUN ./installer --install_pkg
ENTRYPOINT ["tail -f /dev/null"]
