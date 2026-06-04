.PHONY: build test lint

build:
	go build -o cast ./cmd/cast

test:
	go test ./...

lint:
	golangci-lint run
