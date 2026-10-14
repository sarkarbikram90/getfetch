.PHONY: build test run-dry-run run-worker run-api web-dev web-build docker-up docker-down clean

# Build all Go binaries
build:
	@mkdir -p bin
	go build -o bin/worker.exe ./cmd/worker
	go build -o bin/api.exe ./cmd/api

# Run all unit tests
test:
	go test -v ./...

# Run worker in dry-run mode against all seed sources
run-dry-run:
	go run ./cmd/worker -dry-run

# Run worker against single source
crawl-one:
	go run ./cmd/worker -dry-run -source $(SOURCE)

# Run worker with PostgreSQL
run-worker:
	go run ./cmd/worker

# Run REST API server
run-api:
	go run ./cmd/api -port 8080

# Run Next.js frontend dev server
web-dev:
	cd web && npm run dev

# Build Next.js frontend
web-build:
	cd web && npm run build

# Start full containerized stack (Postgres, API, Worker, Web, Prometheus)
docker-up:
	docker compose -f deployments/docker-compose.yml up -d

# Stop containerized stack
docker-down:
	docker compose -f deployments/docker-compose.yml down

# Clean build artifacts
clean:
	rm -rf bin/
