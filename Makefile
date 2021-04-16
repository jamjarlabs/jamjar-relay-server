REGISTRY = jamjarlabs
NAME = jamjar-relay-server
VERSION = latest

LOCAL_PORT=5000
LOCAL_ADDRESS=0.0.0.0

CORS_ORIGINS=http://localhost:8000

default: vendor_modules
	docker build -t $(REGISTRY)/$(NAME):$(VERSION) .

all: linux_amd64 mac_amd64 windows_amd64

package_all:
	rm -f linux_amd64.tar.gz
	rm -f mac_amd64.tar.gz
	rm -f windows_amd64.tar.gz
	tar -czvf linux_amd64.tar.gz dist/linux_amd64/*
	tar -czvf mac_amd64.tar.gz dist/linux_amd64/*
	tar -czvf windows_amd64.tar.gz dist/linux_amd64/*

linux_amd64: vendor_modules
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -o dist/linux_amd64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/linux_amd64/LICENSE

mac_amd64: vendor_modules
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -mod vendor -o dist/mac_amd64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/mac_amd64/LICENSE

windows_amd64: vendor_modules
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -mod vendor -o dist/windows_amd64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/windows_amd64/LICENSE

run: generate vendor_modules
	@echo "=============Running Application Locally============="
	ADDRESS=$(LOCAL_ADDRESS) PORT=$(LOCAL_PORT) CORS_ORIGINS=$(CORS_ORIGINS) go run -mod vendor cmd/jamjar-relay-server/main.go -v 5 -logtostderr true

cli: generate vendor_modules
	go run -mod vendor cmd/cli/main.go ws://$(LOCAL_ADDRESS):$(LOCAL_PORT)/v1/websocket

lint: vendor_modules
	gofmt -s -w .
	go mod tidy
	go list -mod vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

generate:
	protoc -I=specs --go_out=paths=source_relative:./specs $(shell find specs/ -iname "*.proto")

proto:
	./hack/proto.sh

package_proto: proto
	rm -f protobuf.zip
	cd dist/protobuf/ && zip -r ../../protobuf.zip *

vendor_modules:
	go mod vendor
