# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2019 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

PWD := $(shell pwd)
PLATFORM := linux
BINARY := monitor

export GO111MODULE=on

all: test build
deploy: test build

build: clean test cover
	CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=amd64
	go build -a -ldflags '-extldflags "-static"' \
	-o $(PWD)/$(BINARY) cmd/manager/main.go

deploy: build

.PHONY: test
test: clean
	@go test -v ./...

format:
	@go fmt ./...

clean:
	@rm -f $(BINARY)

.PHONY: cover
cover:
	@go test -p 2 ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
