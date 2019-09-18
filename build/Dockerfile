FROM ubuntu:18.04 as base
ADD . /usr/src/multicloud-k8s
USER root
SHELL ["/bin/bash", "-c"]
WORKDIR /usr/src/multicloud-k8s/kud/hosting_providers/containerized
RUN ./installer --install_pkg
ENTRYPOINT ["tail -f /dev/null"]
