MAIN=dp-import-reporter
BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) cmd/$(MAIN)/main.go

.PHONY: debug
debug:
	HUMAN_LOG=1 go run -race cmd/$(MAIN)/main.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: generate
generate:
	go generate -v ./...
