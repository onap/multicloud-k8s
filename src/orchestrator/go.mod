module github.com/onap/multicloud-k8s/src/orchestrator

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/coreos/etcd v3.3.13+incompatible

	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.4.1
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/lib/pq v1.7.0 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/onap/multicloud-k8s/src/clm v0.0.0-00010101000000-000000000000
	github.com/onap/multicloud-k8s/src/rsync v0.0.0-20200630152613-7c20f73e7c5d
	github.com/pkg/errors v0.9.1
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	go.etcd.io/etcd v3.3.12+incompatible
	go.mongodb.org/mongo-driver v1.3.4
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/apimachinery v0.16.9
	k8s.io/helm v2.14.3+incompatible
	k8s.io/kubernetes v1.15.0
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
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

