PROJECT_NAME ?= robot

ROOT_DIR=$(PWD)

TINKOFF_PROTO=$(ROOT_DIR)/investapi
ROBOT_PROTO=$(ROOT_DIR)/robot/proto


all:
	@echo "build			- Build project"
	@echo "setup			- Setup project"
	@echo "setup-dev		- Setup dev dependencies"
	@echo "clean			- Remove compiled proto"
	@echo "compile-proto	- Compile all .proto files"
	@echo "lint				- Run linter"
	@echo "tests			- Run tests"
	@echo "coverage			- Show test's coverage"
	@exit 0

build:
	go build -v ./cmd/run-robot/
	go build -v ./cmd/generate-config/

setup:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	apt-get install -y protobuf-compiler

setup-dev:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

clean:
	rm -f $(TINKOFF_PROTO)/*.go
	rm -f $(ROBOT_PROTO)/*.go
	rm -f ./run-robot
	rm -f ./generate-config
	rm -f ./generated/*

compile-proto:
	make clean
	protoc -I=$(TINKOFF_PROTO) --go_out=$(TINKOFF_PROTO)/ --go-grpc_out=$(TINKOFF_PROTO)/ $(TINKOFF_PROTO)/*

lint:
	go vet ./...
	golangci-lint run

tests:
	go test -race -v -failfast ./...

coverage:
	go test -cover ./...


.PHONY: all build setup setup-dev clean compile-proto lint tests coverage
