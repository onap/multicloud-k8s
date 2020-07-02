FROM golang:1.14.1

WORKDIR /go/src/github.com/onap/multicloud-k8s/src/monitor
COPY ./ ./
RUN make all

FROM ubuntu:16.04

WORKDIR /opt/monitor
RUN groupadd -r monitor && useradd -r -g monitor monitor
RUN chown monitor:monitor /opt/monitor -R
COPY --chown=monitor --from=0 /go/src/github.com/onap/multicloud-k8s/src/monitor/monitor ./

USER monitor
ENTRYPOINT ["/opt/monitor/monitor"]

