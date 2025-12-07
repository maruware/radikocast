.PHONY: clean build test

build:
	go build ./cmd/radikocast/main.go

test:
	go test -v ./...
