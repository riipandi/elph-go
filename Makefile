.DEFAULT_GOAL := help

BINARY_NAME := elph
BUILD_DIR   := ./build/release
GO          := go
GOTEST      := gotestsum
TUI         := $(shell which tui)

# ─── Test Flags ───────────────────────────────────────────────────────────────

TEST_FLAGS := --format short-verbose -- -count=1 -v

.PHONY: build run install
.PHONY: test test-unit test-integration test-coverage
.PHONY: clean deps lint fmt vet views help

# ─── Build ────────────────────────────────────────────────────────────────────

views: ## Regenerate TUI views
	@$(TUI) generate internal/views/

build: views ## Build the application
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd
	@echo "Binary size: $$(du -sh $(BUILD_DIR)/$(BINARY_NAME) | cut -f1) ($$(shasum -a 1 $(BUILD_DIR)/$(BINARY_NAME) | cut -d' ' -f1))"

install: build ## Build and copy binary to ~/.local/bin
	@mkdir -p $(HOME)/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed: $$(command -v $(HOME)/.local/bin/$(BINARY_NAME) 2>/dev/null || echo $(HOME)/.local/bin/$(BINARY_NAME))"

run: views ## Run the application
	@$(GO) run ./cmd

# ─── Testing ──────────────────────────────────────────────────────────────────

test: ## Run all tests
	@$(GOTEST) $(TEST_FLAGS) ./...

test-unit: ## Run unit tests only
	@$(GOTEST) $(TEST_FLAGS) ./internal/adapter/... ./internal/service/... ./internal/handler/...

test-integration: ## Run integration tests
	@$(GOTEST) $(TEST_FLAGS) -tags=integration ./internal/integration/...

test-coverage: ## Run tests with coverage report
	@$(GOTEST) ./internal/... -v -coverprofile=coverage.out
	@$(GO) tool cover -html=coverage.out -o coverage.html

# ─── Code Quality ─────────────────────────────────────────────────────────────

lint: ## Run linter (requires golangci-lint)
	@golangci-lint run ./...

fmt: ## Format code
	@$(GO) fmt ./...

vet: ## Vet code
	@$(GO) vet ./...

# ─── Maintenance ──────────────────────────────────────────────────────────────

deps: ## Download dependencies
	@$(GO) mod download
	@$(GO) mod tidy

clean: ## Clean build artifacts
	@rm -rf $(BUILD_DIR) vendor node_modules coverage.out coverage.html

# ─── Help ─────────────────────────────────────────────────────────────────────

help: ## Show this help
	@printf '\033[33mUsage:\033[0m make \033[36m<target>\033[0m\n'
	@awk -F ':.*## ' '/^[a-zA-Z_-]+:.*## / {printf " \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
