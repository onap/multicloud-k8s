all: clean plugins build tests

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o ./k8plugin ./cmd/main.go

tests:
	go test -v ./... -cover

format:
	go fmt ./...

plugins:
	go build -buildmode=plugin -o ./plugins/deployment/deployment.so ./plugins/deployment/plugin.go
	go build -buildmode=plugin -o ./plugins/namespace/namespace.so ./plugins/namespace/plugin.go
	go build -buildmode=plugin -o ./plugins/service/service.so ./plugins/service/plugin.go
	go build -buildmode=plugin -o ./csar/mock_plugins/mockplugin.so ./csar/mock_plugins/mockplugin.go

clean:
	find . -name "*so" -delete
	@rm -f k8plugin
