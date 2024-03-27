BINARY_NAME=bisk
DIST_DIR=./dist
ARCH_TYPE := $(shell echo "$$(uname -m)-$$(uname -s | awk '{print tolower($$0)}')")
TARGET_PLATFORMS := "aarch64 linux arm64" "arm64 darwin arm64"

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"

all: test build

build:
	@for platform in $(TARGET_PLATFORMS); do \
		alias=$$(echo $$platform | cut -d' ' -f1); \
		os=$$(echo $$platform | cut -d' ' -f2); \
		arch=$$(echo $$platform | cut -d' ' -f3); \
		echo "Building $$alias for $$os-$$arch"; \
		GOOS=$$os GOARCH=$$arch go build -o $(DIST_DIR)/@$$arch-$$os/$(BINARY_NAME) cmd/bisk/main.go; \
	done

clean:
	@echo "Cleaning"
	@GO111MODULE=on go clean
	@rm -rf $(DIST_DIR)

init:
	go mod tidy

run: build
	@echo "Running application..."
	@$(DIST_DIR)/@$(ARCH_TYPE)/$(BINARY_NAME) server start

help:
	@echo "Available commands:"
	@echo "  all    : Run tests and build the project"
	@echo "  build  : Build the project and produce a binary"
	@echo "  test   : Run all the tests"
	@echo "  clean  : Clean the build artifacts"
	@echo "  run    : Run the built binary"

.PHONY: all build test clean run help
