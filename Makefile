.PHONY: build test run-dry-run run-worker docker-up docker-down clean

# Build worker binary
build:
	@mkdir -p bin
	go build -o bin/worker.exe ./cmd/worker

# Run all unit tests
test:
	go test -v -race ./...

# Run worker in dry-run mode against all seed sources
run-dry-run:
	go run ./cmd/worker -dry-run

# Run worker against a specific company board (e.g. make crawl-one SOURCE=gitlab)
crawl-one:
	go run ./cmd/worker -dry-run -source $(SOURCE)

# Start local dependencies (PostgreSQL 16 & Prometheus)
docker-up:
	docker compose -f deployments/docker-compose.yml up -d

# Stop local dependencies
docker-down:
	docker compose -f deployments/docker-compose.yml down

# Clean build artifacts
clean:
	rm -rf bin/
