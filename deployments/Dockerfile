FROM debian:jessie

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY
ENV no_proxy $NO_PROXY

ENV CSAR_DIR "/opt/csar"
ENV KUBE_CONFIG_DIR "/opt/kubeconfig"
ENV DATABASE_TYPE "consul"
ENV DATABASE_IP "127.0.0.1"

EXPOSE 8081

WORKDIR /opt/multicloud/k8s
ADD ./k8plugin ./
ADD ./*.so ./

CMD ["./k8plugin"]
