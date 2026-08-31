.PHONY: all build run test test-race tidy docker-up docker-down clean

# Binary name
BINARY_NAME=bin/taskflow-api

all: test build

build:
	go build -o $(BINARY_NAME) cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

test-race:
	go test -v -race ./...

tidy:
	go mod tidy

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/

