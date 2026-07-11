.PHONY: install dev build dev-fe dev-be build-fe build-be clean help

# ---------- Config ----------
FE_DIR     := web
BUILD_DIR  := build
BE_CMD     := ./cmd/server
BE_BIN     := dbquery-server

# ---------- Install ----------
install: install-fe install-be ## Install all dependencies (FE + BE)

install-fe: ## Install frontend dependencies
	@echo "==> Installing frontend dependencies..."
	cd $(FE_DIR) && npm install

install-be: ## Install backend dependencies
	@echo "==> Installing backend dependencies..."
	go mod download

# ---------- Dev ----------
dev: ## Run FE and BE concurrently (Ctrl+C kills both)
	@echo "==> Starting frontend and backend concurrently..."
	@trap 'kill 0' INT TERM; \
	cd $(FE_DIR) && npm run dev & \
	if command -v air >/dev/null 2>&1; then \
		echo "    Using air for hot-reload..."; \
		air & \
	else \
		echo "    'air' not found — falling back to 'go run' (no hot-reload)."; \
		echo "    Install air: go install github.com/air-verse/air@latest"; \
		go run $(BE_CMD) & \
	fi; \
	wait

dev-fe: ## Start frontend dev server only (Vite, watch mode)
	@echo "==> Starting frontend dev server..."
	cd $(FE_DIR) && npm run dev

dev-be: ## Start backend API server only (air or go run)
	@echo "==> Starting backend API server..."
	@if command -v air >/dev/null 2>&1; then \
		echo "    Using air for hot-reload..."; \
		air; \
	else \
		echo "    'air' not found — falling back to 'go run' (no hot-reload)."; \
		echo "    Install air: go install github.com/air-verse/air@latest"; \
		go run $(BE_CMD); \
	fi

# ---------- Build ----------
build: build-fe build-be ## Build frontend and backend into build/
	@echo ""
	@echo "==> Build complete!"
	@echo "    Binary:  $(BUILD_DIR)/$(BE_BIN)"
	@echo "    Frontend: $(BUILD_DIR)/web/"
	@echo ""
	@echo "    To run:  cd $(BUILD_DIR) && ./$(BE_BIN)"

build-fe: ## Build frontend for production into build/web/
	@echo "==> Building frontend..."
	cd $(FE_DIR) && npm run build
	@echo "==> Copying frontend build to $(BUILD_DIR)/web/..."
	mkdir -p $(BUILD_DIR)/web
	cp -r $(FE_DIR)/dist/* $(BUILD_DIR)/web/

build-be: ## Build backend binary into build/
	@echo "==> Building backend..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BE_BIN) $(BE_CMD)

# ---------- Clean ----------
clean: ## Remove build artifacts
	@echo "==> Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(FE_DIR)/dist

# ---------- Help ----------
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
