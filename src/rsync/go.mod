module github.com/onap/multicloud-k8s/src/rsync

go 1.13

require (
	github.com/golang/protobuf v1.4.1
	github.com/googleapis/gnostic v0.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.5 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onap/multicloud-k8s/src/clm v0.0.0-00010101000000-000000000000
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200601021239-7959bd4c6fd4
	github.com/pkg/errors v0.8.1
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.27.1
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kubernetes v1.14.1
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
	github.com/onap/multicloud-k8s/src/orchestrator => ../orchestrator
	github.com/onap/multicloud-k8s/src/rsync => ../rsync
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)
