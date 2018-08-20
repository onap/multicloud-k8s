GOPATH := $(GOPATH)

export GOPATH ...

.DEFAULT_GOAL := build

build: check_gopath plugins run_tests
deploy: check_gopath plugins generate_binary run_tests

generate_binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o $(GOPATH)/target/k8plugin $(GOPATH)/src/k8-plugin-multicloud/cmd/main.go

run_tests:
	cd $(GOPATH)/src/k8-plugin-multicloud && go test -v ./... -cover

format:
	cd $(GOPATH)/src/k8-plugin-multicloud && go fmt ./...

plugins:
	go build -buildmode=plugin -o $(GOPATH)/src/k8-plugin-multicloud/plugins/deployment/deployment.so $(GOPATH)/src/k8-plugin-multicloud/plugins/deployment/plugin.go
	go build -buildmode=plugin -o $(GOPATH)/src/k8-plugin-multicloud/plugins/namespace/namespace.so $(GOPATH)/src/k8-plugin-multicloud/plugins/namespace/plugin.go
	go build -buildmode=plugin -o $(GOPATH)/src/k8-plugin-multicloud/plugins/service/service.so $(GOPATH)/src/k8-plugin-multicloud/plugins/service/plugin.go
	go build -buildmode=plugin -o $(GOPATH)/src/k8-plugin-multicloud/csar/mock_plugins/mockplugin.so $(GOPATH)/src/k8-plugin-multicloud/csar/mock_plugins/mockplugin.go

check_gopath:
ifndef GOPATH
  $(error GOPATH is not set)
endif
