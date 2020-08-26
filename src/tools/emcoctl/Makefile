# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

export GO111MODULE=on

all: clean
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64
	@go build  -o ./emcoctl ./emcoctl.go

build: clean test cover
deploy: build

.PHONY: test
test: clean
	@go test -race ./...

format:
	@go fmt ./...

clean:
	@rm -f emcoctl coverage.html coverage.out

.PHONY: cover
cover:
	@go test -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
