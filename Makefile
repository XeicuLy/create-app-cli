.PHONY: setup build test lint

setup:
	mise install
	mise exec -- go mod download

build:
	mise exec -- go build -o xeikit ./cmd/xeikit

test:
	mise exec -- go test ./... -race -v

lint:
	mise exec -- golangci-lint run
