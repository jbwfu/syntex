# Makefile for the syntex project

# Variables
BINARY_NAME=syntex
CMD_PATH=./cmd/syntex
GO_EXEC ?= $(shell which go)

# Installation directories
# Users can override these from the command line, e.g., `make install PREFIX=/usr`
PREFIX ?= /usr/local
DESTDIR ?= $(PREFIX)/bin

# Default target executed when you run `make`
.DEFAULT_GOAL := build

##@ Build

build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@$(GO_EXEC) build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully."

##@ Run

run: ## Run the application with arguments (e.g., make run ARGS="path/to/dir")
	@$(GO_EXEC) run $(CMD_PATH) $(ARGS)

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


.PHONY: build run clean install uninstall
