.PHONY: setup build test lint

setup:
	mise install
	go mod download

build:
	go build -o xeikit ./cmd/xeikit

test:
	go test ./... -race -v

lint:
	golangci-lint run
