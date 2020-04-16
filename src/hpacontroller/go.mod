module github.com/onap/multicloud-k8s/src/hpacontroller

go 1.13

require (
	github.com/golang/protobuf v1.3.4
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200324023056-8c6fb0cbc3ce
	github.com/pkg/errors v0.9.1
	google.golang.org/grpc v1.28.0
)

replace github.com/onap/multicloud-k8s/src/orchestrator => /go/src/src/orchestrator
