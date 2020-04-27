module rsync

go 1.13

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/engine v1.13.1
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.3.4
	github.com/google/uuid v1.1.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.5 // indirect
	github.com/hashicorp/consul/api v1.4.0
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/lib/pq v1.6.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onap/multicloud-k8s/src/ncm v0.0.0-20200515060444-c77850a75eee
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200527175204-ef27eb4d63f1
	github.com/onap/multicloud-k8s/src/rsync v0.0.0-20200529003854-0a7bf256bde5
	github.com/pkg/errors v0.8.1
	github.com/rubenv/sql-migrate v0.0.0-20200429072036-ae26b214fa43 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	go.etcd.io/etcd v3.3.12+incompatible
	go.mongodb.org/mongo-driver v1.0.0
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	google.golang.org/grpc v1.27.1
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	k8s.io/api v0.0.0-20190831074750-7364b6bdad65
	k8s.io/apiextensions-apiserver v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381
	k8s.io/apiserver v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/cli-runtime v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/helm v2.14.3+incompatible
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
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
