# Makefile for API Gateway

.PHONY: build clean run test deps

build:
	go build -o bin/api-gateway cmd/server/main.go

build-static:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o bin/api-gateway-static cmd/server/main.go

clean:
	rm -rf bin/

run:
	./bin/api-gateway

test:
	go test ./...

deps:
	go mod download
	go mod tidy

install:
	go install ./cmd/server

.DEFAULT_GOAL := build