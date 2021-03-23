LOCAL_PORT=8000
LOCAL_ADDRESS=0.0.0.0

vendor_modules:
	go mod vendor

run: generate vendor_modules
	@echo "=============Running Application Locally============="
	ADDRESS=$(LOCAL_ADDRESS) PORT=$(LOCAL_PORT) go run -mod vendor cmd/jamjar-relay-server/main.go -v 5 -logtostderr true

cli: generate vendor_modules
	go run -mod vendor cmd/cli/main.go ws://localhost:8000/v1/websocket

lint:
	gofmt -s -w .

generate:
	protoc -I=specs --go_out=paths=source_relative:./specs $(shell find specs/ -iname "*.proto")
