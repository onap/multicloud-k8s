module github.com/onap/multicloud-k8s/src/monitor

require (
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/onap/multicloud-k8s/src/k8splugin v0.0.0-20190808131943-845cdd2aa5d7 // indirect
	github.com/operator-framework/operator-sdk v0.9.1-0.20190729152335-7a35cfc9a7cf
	github.com/operator-framework/operator-sdk-samples v0.0.0-20190529081445-bd30254f3a7e
	github.com/pkg/errors v0.8.1
	github.com/spf13/pflag v1.0.3
	k8s.io/api v0.0.0-20190814101207-0772a1bdf941
	k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery v0.0.0-20190814100815-533d101be9a6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/sample-controller v0.0.0-20190814141925-f27ac7da6c3e // indirect
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
replace bitbucket.org/ww/goautoneg => github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.9.0

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

go 1.13
