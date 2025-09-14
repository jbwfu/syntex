# Makefile for the syntex project

# Variables
BINARY_NAME=syntex
CMD_PATH=./cmd/syntex
GO_EXEC ?= $(shell which go)

# Installation directories
# Users can override these from the command line, e.g., `make install PREFIX=/usr`
PREFIX ?= /usr/local
DESTDIR ?= $(PREFIX)/bin

# Versioning information for ldflags
GIT_TAG_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
VERSION         := $(if $(GIT_TAG_VERSION),$(GIT_TAG_VERSION),dev)
# Get short commit hash, fallback to "unknown".
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Get current UTC build date/time in ISO 8601 format, fallback to "unknown".
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")

LDFLAGS := -ldflags "\
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.buildDate=$(BUILD_DATE)' \
	-s -w \
	"

# Default target executed when you run `make`
.DEFAULT_GOAL := build

##@ Build

build: ## Build the application binary with embedded version info
	@echo "Building $(BINARY_NAME) (Version: $(VERSION), Commit: $(COMMIT), Date: $(BUILD_DATE))..."
	@$(GO_EXEC) build $(LDFLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully."

##@ Run

run: ## Run the application with arguments and embedded version info (e.g., make run ARGS="path/to/dir")
	@echo "Running $(BINARY_NAME) (Version: $(VERSION), Commit: $(COMMIT), Date: $(BUILD_DATE))..."
	@$(GO_EXEC) run $(LDFLAGS) $(CMD_PATH) $(ARGS)

##@ Clean

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

##@ Install

install: build ## Install the application to DESTDIR (default: /usr/local/bin)
	@echo "Installing $(BINARY_NAME) to $(DESTDIR)..."
	@install -d -m 0755 "$(DESTDIR)"
	@install -m 0755 "$(BINARY_NAME)" "$(DESTDIR)"
	@echo "$(BINARY_NAME) installed successfully."

##@ Uninstall

uninstall: ## Uninstall the application from DESTDIR
	@echo "Uninstalling $(BINARY_NAME) from $(DESTDIR)..."
	@rm -f "$(DESTDIR)/$(BINARY_NAME)"
	@echo "$(BINARY_NAME) uninstalled."

##@ Help

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'


.PHONY: build run clean install uninstall help
