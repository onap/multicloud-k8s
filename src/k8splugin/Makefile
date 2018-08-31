# Copyright 2018 Intel Corporation.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GOPATH := $(shell realpath "$(PWD)/../../")
DEPENDENCIES := github.com/golang/dep/cmd/dep

export GOPATH ...

.PHONY: plugins

build: clean dep plugins tests
deploy: clean dep plugins build_binary tests

build_binary:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o ./k8plugin ./cmd/main.go

tests:
	go test -v ./... -cover

format:
	go fmt ./...

plugins:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=plugin -a -tags netgo -o ./plugins/deployment/deployment.so ./plugins/deployment/plugin.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=plugin -a -tags netgo -o ./plugins/namespace/namespace.so ./plugins/namespace/plugin.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=plugin -a -tags netgo -o ./plugins/service/service.so ./plugins/service/plugin.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=plugin -a -tags netgo -o ./mock_files/mock_plugins/mockplugin.so ./mock_files/mock_plugins/mockplugin.go

dep:
	go get -u $(DEPENDENCIES)
	$(GOPATH)/bin/dep ensure

clean:
	find . -name "*so" -delete
	@rm -f k8plugin
