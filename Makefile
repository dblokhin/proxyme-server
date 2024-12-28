ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

GO ?= go
GOBUILD = $(GO) build
GOCLEAN = $(GO) clean
GOTEST = $(GO) test
GOGET = $(GO) get
BINARY_NAME = proxyme
GOLANGCI_LINT_VERSION := v1.60.2
BIN_DIR := $(shell go env GOPATH)/bin

build:
	$(GOBUILD) -o $(BINARY_NAME) .

run:
	$(GO) run .

test:
	$(GOTEST) -cover -count=1 ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...
	$(BIN_DIR)/golangci-lint run ./...

deps:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)
	@go install github.com/mattn/goveralls@latest

cover:
	goveralls

docker-pub:
	docker buildx build --platform linux/amd64,linux/arm64 -t docker.io/dblokhin/$(BINARY_NAME):latest -t docker.io/dblokhin/$(BINARY_NAME):$(shell git describe --tags --abbrev=0) --push .

docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run --rm -it -p 1080:1080 -e PROXY_NOAUTH=yes $(BINARY_NAME)

.PHONY: build clean test run fmt lint deps cover docker-pub docker-build docker-run
