# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

export GO111MODULE=on
GO ?= go

.PHONY: plugins


all: clean plugins
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64
	@$(GO) build -tags netgo -o ./k8plugin ./cmd/main.go

# The following is done this way as each patch on CI runs build and each merge runs deploy. So for build we don't need to build binary and hence
# no need to create a static binary with additional flags. However, for generating binary, additional build flags are necessary. This if used with
# mock plugin errors out for unit tests. So the seperation avoids the error.

build: clean plugins test cover
deploy: build

.PHONY: test
test: clean
	@$(GO) build -race -buildmode=plugin -o ./mock_files/mock_plugins/mockplugin.so ./mock_files/mock_plugins/mockplugin.go
	@$(GO) test -race ./...

format:
	@$(GO) fmt ./...

plugins:
	@find plugins -maxdepth 1 -type d -not -path plugins -exec sh -c "ls {}/plugin.go | xargs $(GO) build -buildmode=plugin -tags netgo -o $(basename {}).so" \;

clean:
	@find . -name "*so" -delete
	@rm -f k8plugin

.PHONY: cover
cover:
	@$(GO) test -race ./... -coverprofile=coverage.out
	@$(GO) tool cover -html=coverage.out -o coverage.html
