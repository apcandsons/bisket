BINARY_NAME=bisq
DIST_DIR=./dist
ARCH_TYPE := $(shell uname -m)

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"

all: test build

build:
	@echo "Building for $(ARCH_TYPE)"
	@GO112MODULE=on go build -o $(DIST_DIR)/@$(ARCH_TYPE)/$(BINARY_NAME) ./cmd/bisq

clean:
	@echo "Cleaning"
	@GO111MODULE=on go clean
	@rm -rf $(DIST_DIR)

init:
	go mod tidy

run: build
	@echo "Running application..."
	@./$(DIST_DIR)/@$(ARCH_TYPE)/$(BINARY_NAME) server start

help:
	@echo "Available commands:"
	@echo "  all    : Run tests and build the project"
	@echo "  build  : Build the project and produce a binary"
	@echo "  test   : Run all the tests"
	@echo "  clean  : Clean the build artifacts"
	@echo "  run    : Run the built binary"

.PHONY: all build test clean run help
