# NOTE: Possibly only necessary with versions below 1.16?
# export GO111MODULE=on

export PATH := $(GOPATH)/bin:$(PATH)

BINARY_VERSION?=0.0.1
BINARY_OUTPUT?=rustbot
EXTRA_FLAGS?=-mod=vendor

RUSTPLUS_DIR?=$(CURDIR)/rustplus
RUSTPLUS_PROTO_PATH?=$(RUSTPLUS_DIR)/rustplus.proto

.PHONY: all install build test clean deps upgrade

all: clean deps build

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(RUSTPLUS_DIR)/rustplus.pb.go

# TODO: Don't mark this a phony, but instead see if rustplus.proto has changed?
protogen: $(RUSTPLUS_PROTO_PATH)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	protoc --go_out=$(RUSTPLUS_DIR) --go_opt=paths=source_relative --proto_path=$(RUSTPLUS_DIR) $(RUSTPLUS_PROTO_PATH)

deps: protogen
	go build -v $(EXTRA_FLAGS) ./...

build: deps
	go build -v $(EXTRA_FLAGS) -ldflags "-X main.Version=$(BINARY_VERSION)" -o $(BINARY_OUTPUT)

test: build
	go test -v $(EXTRA_FLAGS) -race -coverprofile=coverage.txt -covermode=atomic ./...

install: build
	go install -v $(EXTRA_FLAGS) -ldflags "-X main.Version=$(BINARY_VERSION)"

upgrade: deps
	go get -u ./...
	go mod vendor

tidy: deps
	go mod tidy
