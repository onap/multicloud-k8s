module github.com/onap/multicloud-k8s/src/monitor

go 1.17

require (
	github.com/go-openapi/spec v0.19.4
	github.com/operator-framework/operator-sdk v0.19.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	sigs.k8s.io/controller-runtime v0.6.0
)

// Pinned to kubernetes-1.13.4
replace (
	k8s.io/api => k8s.io/api v0.16.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.9
	k8s.io/client-go => k8s.io/client-go v0.16.9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.11
)
