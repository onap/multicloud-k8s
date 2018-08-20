GOPATH := $(GOPATH)

export GOPATH ...

.DEFAULT_GOAL := build

.PHONY: plugins

build: check_gopath plugins run_tests
deploy: check_gopath plugins generate_binary run_tests

generate_binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o $(GOPATH)/target/k8plugin $(GOPATH)/src/k8s/cmd/main.go

run_tests:
	cd $(GOPATH)/src/k8s && go test -v ./... -cover

format:
	cd $(GOPATH)/src/k8s && go fmt ./...

plugins:
	go build -buildmode=plugin -o $(GOPATH)/src/k8s/plugins/deployment/deployment.so $(GOPATH)/src/k8s/plugins/deployment/plugin.go
	go build -buildmode=plugin -o $(GOPATH)/src/k8s/plugins/namespace/namespace.so $(GOPATH)/src/k8s/plugins/namespace/plugin.go
	go build -buildmode=plugin -o $(GOPATH)/src/k8s/plugins/service/service.so $(GOPATH)/src/k8s/plugins/service/plugin.go
	go build -buildmode=plugin -o $(GOPATH)/src/k8s/csar/mock_plugins/mockplugin.so $(GOPATH)/src/k8s/csar/mock_plugins/mockplugin.go

check_gopath:
ifndef GOPATH
  $(error GOPATH is not set)
endif
