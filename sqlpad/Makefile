.PHONY: install dev build dev-fe dev-be build-fe build-be clean help

# ---------- Config ----------
FE_DIR    := web
BE_CMD    := ./cmd/server
BE_BIN    := sqlpad-server

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
build: build-fe build-be ## Build frontend and backend

build-fe: ## Build frontend for production
	@echo "==> Building frontend..."
	cd $(FE_DIR) && npm run build

build-be: ## Build backend binary
	@echo "==> Building backend..."
	go build -o $(BE_BIN) $(BE_CMD)

# ---------- Clean ----------
clean: ## Remove build artifacts
	@echo "==> Cleaning..."
	rm -rf $(FE_DIR)/dist
	rm -f $(BE_BIN)

# ---------- Help ----------
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
