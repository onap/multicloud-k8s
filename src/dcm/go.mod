module github.com/onap/multicloud-k8s/src/dcm

require (
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/onap/multicloud-k8s/src/clm v0.0.0-20200630152613-7c20f73e7c5d
	github.com/onap/multicloud-k8s/src/monitor v0.0.0-20200818155723-a5ffa8aadf49
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200818155723-a5ffa8aadf49
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/stretchr/testify v1.5.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/kubernetes v1.16.9
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
	github.com/onap/multicloud-k8s/src/monitor => ../monitor
	github.com/onap/multicloud-k8s/src/orchestrator => ../orchestrator
	k8s.io/api => k8s.io/api v0.16.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.10-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.16.9
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.9
	k8s.io/client-go => k8s.io/client-go v0.16.9
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.9
	k8s.io/code-generator => k8s.io/code-generator v0.16.10-beta.0
	k8s.io/component-base => k8s.io/component-base v0.16.9
	k8s.io/cri-api => k8s.io/cri-api v0.16.13-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.9
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.9
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.9
	k8s.io/kubectl => k8s.io/kubectl v0.16.9
	k8s.io/kubelet => k8s.io/kubelet v0.16.9
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.16.9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.9
	k8s.io/metrics => k8s.io/metrics v0.16.9
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.9
)

go 1.13
