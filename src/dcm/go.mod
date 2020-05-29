module github.com/onap/multicloud-k8s/src/dcm

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gorilla/handlers v1.3.0
	github.com/gorilla/mux v1.7.3
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/lib/pq v1.5.2 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/onap/multicloud-k8s/src/orchestrator v0.0.0-20200527175204-ef27eb4d63f1
	github.com/pkg/errors v0.8.1
	github.com/rubenv/sql-migrate v0.0.0-20200429072036-ae26b214fa43 // indirect
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apiextensions-apiserver v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/apiserver v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/cli-runtime v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/cloud-provider v0.0.0-00010101000000-000000000000 // indirect
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

go 1.13
