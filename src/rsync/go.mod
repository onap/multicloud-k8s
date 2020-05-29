module github.com/onap/multicloud-k8s/src/rsync

go 1.13

require (
	github.com/golang/protobuf v1.3.4
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200527175204-ef27eb4d63f1
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/grpc v1.27.1
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
)

replace (
	github.com/onap/multicloud-k8s/src/ncm => ../ncm
	github.com/onap/multicloud-k8s/src/orchestrator => ../orchestrator
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)
