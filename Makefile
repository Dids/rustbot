# NOTE: Only necessary with versions below 1.16
# export GO111MODULE=on

export PATH := $(GOPATH)/bin:$(PATH)

BINARY_VERSION?=0.0.1
BINARY_OUTPUT?=rustbot
EXTRA_FLAGS?=-mod=vendor

.PHONY: all install build test clean deps upgrade

all: deps build

install:
	go install -v $(EXTRA_FLAGS) -ldflags "-X main.Version=$(BINARY_VERSION)"

build:
	go build -v $(EXTRA_FLAGS) -ldflags "-X main.Version=$(BINARY_VERSION)" -o $(BINARY_OUTPUT)

test:
	go test -v $(EXTRA_FLAGS) -race -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

deps:
	go build -v $(EXTRA_FLAGS) ./...

tidy:
	go mod tidy

upgrade:
	go get -u ./...
	go mod vendor

protogen:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	protoc --go_out=$(CURDIR)/rustplus --go_opt=paths=source_relative --proto_path=$(CURDIR)/rustplus $(CURDIR)/rustplus/rustplus.proto
