module github.com/onap/multicloud-k8s/src/monitor

go 1.14

require (
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/go-openapi/spec v0.19.2
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/operator-framework/operator-sdk v0.9.1-0.20190729152335-7a35cfc9a7cf
	github.com/spf13/pflag v1.0.3
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc // indirect
	k8s.io/api v0.0.0-20190814101207-0772a1bdf941
	k8s.io/apimachinery v0.0.0-20190814100815-533d101be9a6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20190813173942-955ffa8fcfc9 // indirect
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190709113604-33be087ad058
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.10
)

// Pinned to kubernetes-1.13.4
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v1.6.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde
)

// Remove hg dependency using this mirror
replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.9.0
