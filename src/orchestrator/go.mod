module github.com/onap/multicloud-k8s/src/orchestrator

require (
	github.com/coreos/etcd v3.3.12+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.4
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.6.2
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/onap/multicloud-k8s/src/k8splugin v0.0.0-20191115005109-f168ebb73d8d // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	go.etcd.io/etcd v3.3.12+incompatible
	go.mongodb.org/mongo-driver v1.0.0
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	google.golang.org/genproto v0.0.0-20200305110556-506484158171 // indirect
	google.golang.org/grpc v1.27.1
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/helm v2.14.3+incompatible
	k8s.io/klog v1.0.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)

go 1.13
