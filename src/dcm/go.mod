module github.com/onap/multicloud-k8s/src/dcm

require (
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200601021239-7959bd4c6fd4
	github.com/pkg/errors v0.8.1
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.8
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
	github.com/onap/multicloud-k8s/src/orchestrator => ../orchestrator
	github.com/onap/multicloud-k8s/src/monitor => ../monitor
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)

go 1.13
