GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

all: install build-static-amd64 build-static-arm64

build: go.sum
	@echo "building cosmos-node-finder"
	@go build -mod=readonly -o build/cosmos-node-finder .

install: go.sum
	@echo "installing cosmos-node-finder"
	@go build -mod=readonly -o $(GOBIN)/cosmos-node-finder .

build-static: build-static-amd64 build-static-arm64

build-static-amd64:
	@echo "building cosmos-node-finder amd64 static binary..."
	@GOOS=linux GOARCH=amd64 go build -mod=readonly -o build/cosmos-node-finder-amd64 -a -tags netgo -ldflags ' -extldflags "-static"' .

build-static-arm64:
	@echo "building cosmos-node-finder arm64 static binary..."
	@GOOS=linux GOARCH=arm64 go build -mod=readonly -o build/cosmos-node-finder-arm64 -a -tags netgo -ldflags '-extldflags "-static"' .

.PHONY: all build build-static-amd64 build-static-arm64
