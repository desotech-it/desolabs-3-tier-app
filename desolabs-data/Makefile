GOCMD ?= go
GOFMT ?= gofmt

NAME := desolabs-data
FULLCOMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
TAG := $(shell git describe --tags --dirty 2>/dev/null || echo "v0.0.0-dev")
VERSION := $(TAG:v%=%)

GOOS := $(shell $(GOCMD) env GOOS)
GOARCH := $(shell $(GOCMD) env GOARCH)

OUTPUT_BIN_DIR = bin

DOCKER_REGISTRY ?= r.deso.tech
DOCKER_PROJECT ?= desolabs
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_PROJECT)/data

LDFLAGS = -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(FULLCOMMIT)' \
	-X 'main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'

BUILD_FLAGS = -trimpath -ldflags "$(LDFLAGS)" -o '$(OUTPUT_BIN_DIR)/$(NAME)'

.DEFAULT_GOAL := build

.PHONY: build
build:
	$(GOCMD) build $(BUILD_FLAGS) .

.PHONY: clean
clean:
	$(GOCMD) clean
	rm -rf '$(OUTPUT_BIN_DIR)'

.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

.PHONY: test
test:
	$(GOCMD) test -v ./...

.PHONY: docker
docker:
	docker buildx build --push \
		--platform linux/amd64,linux/arm64 \
		-t '$(DOCKER_IMAGE):$(VERSION)' \
		-t '$(DOCKER_IMAGE):latest' \
		.
