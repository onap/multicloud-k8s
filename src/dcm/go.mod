module github.com/onap/multicloud-k8s/src/dcm

require (
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200601021239-7959bd4c6fd4
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.8
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
	github.com/onap/multicloud-k8s/src/orchestrator => ../orchestrator

	// k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	// k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	// k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	// k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	// k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	// k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	// k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d

	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e

)

go 1.13

