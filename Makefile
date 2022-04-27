PROJECT_NAME ?= robot
TINKOFF_PROTO=$(PWD)/investapi/contract
ROBOT_PROTO=$(PWD)/robot/proto

all:
	@echo "clean			- Remove compiled proto"
	@echo "lint				- Run linter"
	@echo "coverage			- Show tests coverage"
	@echo "tests			- Run tests"
	@exit 0

setup:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

clean:
	rm -f $(TINKOFF_PROTO)/*.go
	rm -f $(ROBOT_PROTO)/*.go

compile-proto:
	make clean
	protoc -I=$(TINKOFF_PROTO) --go_out=$(TINKOFF_PROTO)/ $(TINKOFF_PROTO)/*
	protoc -I=$(ROBOT_PROTO) --go_out=$(ROBOT_PROTO)/ $(ROBOT_PROTO)/*

lint:
	golangci-lint run

tests:
	go test -v -failfast ./robot

coverage:
	go test -cover ./robot

.PHONY: setup clean test lint coverage