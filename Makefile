BINARY_NAME=ivory-secp256k1
BUILD_VERSION=0.1.0
VGO=go
# Use the go command to build the binary
# Ensure that the go command is available in your PATH
# If you are using Go modules, you can use 'go mod tidy' to manage dependencies
# and 'go build' to compile the binary.
# If you are using a specific version of Go, you can set the GO environment variable
# to point to the desired version, e.g., export GO=/usr/local/go/bin/go
SRC_GOFILES := $(shell find . -name '*.go' -print)
.DELETE_ON_ERROR:

all: build test
test: deps
		$(VGO) test  ./... -cover -coverprofile=coverage.txt -covermode=atomic
ethsign: ${SRC_GOFILES}
		$(VGO) build -o ${BINARY_NAME} -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=$(BUILD_VERSION)" -tags=prod -v
build: ethsign
clean: build
		$(VGO) clean
		rm -f ${BINARY_NAME}
deps:
		$(VGO) get
