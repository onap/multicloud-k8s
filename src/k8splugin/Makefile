# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

GOPATH := $(shell realpath "$(PWD)/../../")
DEPENDENCIES := github.com/golang/dep/cmd/dep

export GOPATH ...

.PHONY: plugins


all: clean dep
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64
	@find plugins -type d -not -path plugins -exec sh -c "ls {}/plugin.go | xargs go build -buildmode=plugin -a -tags netgo -o $(basename {}).so" \;
	@go build -a -tags netgo -o ./k8plugin ./cmd/main.go

# The following is done this way as each patch on CI runs build and each merge runs deploy. So for build we don't need to build binary and hence
# no need to create a static binary with additional flags. However, for generating binary, additional build flags are necessary. This if used with
# mock plugin errors out for unit tests. So the seperation avoids the error.

build: clean dep plugins test
deploy: build

.PHONY: test
test: unit integration

.PHONY: unit
unit:
	@go test -v -tags 'unit' ./...

.PHONY: integration
integration: clean
	@go build -buildmode=plugin -o ./mock_files/mock_plugins/mockplugin.so ./mock_files/mock_plugins/mockplugin.go
	@go test -v -tags 'integration' ./...

.PHONY: format
format:
	@go fmt ./...

.PHONY: plugins
plugins:
	@find plugins -type d -not -path plugins -exec sh -c "ls {}/plugin.go | xargs go build -buildmode=plugin -o $(basename {}).so" \;

.PHONY: dep
dep:
	@go get -u $(DEPENDENCIES)
	$(GOPATH)/bin/dep ensure

.PHONY: clean
clean:
	find . -name "*so" -delete
	@rm -f k8plugin
