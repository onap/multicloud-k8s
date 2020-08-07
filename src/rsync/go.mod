module github.com/onap/multicloud-k8s/src/rsync

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.1
	github.com/googleapis/gnostic v0.4.0
	github.com/jonboulle/clockwork v0.1.0
	github.com/onap/multicloud-k8s/src/clm v0.0.0-00010101000000-000000000000
	github.com/onap/multicloud-k8s/src/monitor v0.0.0-20200818155723-a5ffa8aadf49
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200818155723-a5ffa8aadf49
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.28.0
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/cli-runtime v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	k8s.io/kubectl v0.18.2
	k8s.io/kubernetes v1.14.1
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
	github.com/onap/multicloud-k8s/src/monitor => ../monitor
	github.com/onap/multicloud-k8s/src/orchestrator => ../orchestrator
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/kubectl => k8s.io/kubectl v0.17.3
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.1
)
