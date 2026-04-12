.PHONY: dev dev-frontend dev-backend build test lint clean

# Port used by the backend dev server. Override with: make dev BACKEND_PORT=19000
BACKEND_PORT ?= 18800

# Run both frontend and backend dev servers
dev:
	@echo "Freeing port $(BACKEND_PORT)..."
	@lsof -ti :$(BACKEND_PORT) | xargs kill -9 2>/dev/null || true
	@echo "Starting backend and frontend dev servers..."
	@$(MAKE) dev-backend & $(MAKE) dev-frontend BACKEND_PORT=$(BACKEND_PORT)

# Start frontend dev server (Vite, with proxy to backend)
dev-frontend:
	cd frontend && VITE_API_PROXY_TARGET=http://localhost:$(BACKEND_PORT) bun install && bun run dev

# Start backend dev server without embedded UI to avoid stale frontend assets
dev-backend:
	@$(MAKE) -C .. build
	cd backend && JANE_AI_BINARY="$(abspath ../build/jane-ai)" JANE_AI_LAUNCHER_UI_MODE=api-only go run . --port $(BACKEND_PORT)

# Build frontend and embed into Go binary
build:
	cd frontend && bun install && bun run build:backend
	cd backend && go build -o picoclaw-web .

# Run all tests
test:
	cd backend && go test ./...
	cd frontend && bun install && bun run lint

# Lint and format
lint:
	cd backend && go vet ./...
	cd frontend && bun install && bun run check

# Clean build artifacts
clean:
	rm -rf frontend/dist backend/dist backend/picoclaw-web
	mkdir -p backend/dist && touch backend/dist/.gitkeep
