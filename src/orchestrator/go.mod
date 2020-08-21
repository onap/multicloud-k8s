module github.com/onap/multicloud-k8s/src/orchestrator

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1
	github.com/coreos/etcd v3.3.12+incompatible
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.1
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/lib/pq v1.6.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/onap/multicloud-k8s/src/monitor v0.0.0-20200630152613-7c20f73e7c5d
	github.com/onap/multicloud-k8s/src/ncm v0.0.0-20200515060444-c77850a75eee
	github.com/onap/multicloud-k8s/src/rsync v0.0.0-20200630152613-7c20f73e7c5d
	github.com/pkg/errors v0.9.1
	github.com/rubenv/sql-migrate v0.0.0-20200429072036-ae26b214fa43 // indirect
	github.com/russross/blackfriday v1.5.2
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	go.etcd.io/etcd v3.3.12+incompatible
	go.mongodb.org/mongo-driver v1.1.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	google.golang.org/grpc v1.28.0
	google.golang.org/protobuf v1.24.0
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 v3.0.0-20200506231410-2ff61e1afc86
	k8s.io/apimachinery v0.18.2
	k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d // indirect
	k8s.io/helm v2.14.3+incompatible
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
)

replace (
	github.com/onap/multicloud-k8s/src/clm => ../clm
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
