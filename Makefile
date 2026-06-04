.PHONY: build test

build:
	go build -o cast ./cmd/cast

test:
	go test ./...
