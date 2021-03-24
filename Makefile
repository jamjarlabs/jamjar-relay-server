REGISTRY = jamjarlabs
NAME = jamjar-relay-server
VERSION = latest

LOCAL_PORT=8000
LOCAL_ADDRESS=0.0.0.0

default: vendor_modules
	docker build -t $(REGISTRY)/$(NAME):$(VERSION) .

linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -o dist/linux_amd64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/linux_amd64/LICENSE

mac_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -mod vendor -o dist/mac_amd64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/mac_amd64/LICENSE

mac_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -mod vendor -o dist/mac_arm64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/mac_arm64/LICENSE

windows_amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -mod vendor -o dist/windows_amd64/$(NAME) ./cmd/jamjar-relay-server
	cp LICENSE dist/windows_amd64/LICENSE

run: generate vendor_modules
	@echo "=============Running Application Locally============="
	ADDRESS=$(LOCAL_ADDRESS) PORT=$(LOCAL_PORT) go run -mod vendor cmd/jamjar-relay-server/main.go -v 5 -logtostderr true

cli: generate vendor_modules
	go run -mod vendor cmd/cli/main.go ws://$(LOCAL_ADDRESS):$(LOCAL_PORT)/v1/websocket

lint: vendor_modules
	gofmt -s -w .
	go mod tidy
	go list -mod vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

generate:
	protoc -I=specs --go_out=paths=source_relative:./specs $(shell find specs/ -iname "*.proto")

vendor_modules:
	go mod vendor
