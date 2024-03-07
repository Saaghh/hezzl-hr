tidy:
	go mod tidy

fmt:
	gofumpt -w .
	gci write . --skip-generated -s standard -s default

lint: tidy fmt build
	golangci-lint run

serve: up
	go run ./cmd/apiserver

up:
	docker compose up -d

build:
	go build -o ./bin/apiserver ./cmd/apiserver

test: build up
	go test -v ./tests

.PHONY: build tidy fmt lint serve up test

.DEFAULT_GOAL := lint
