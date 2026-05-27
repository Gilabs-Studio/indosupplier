.PHONY: build run dev test clean docker-build docker-run docker-up docker-down migrate seed deps analyze

# Build the application
build:
	go build -o server ./cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run with hot reload (requires: go install github.com/air-verse/air@latest)
# Prerequisites: docker compose up -d postgres redis
dev:
	air

# Run migrations (explicit, does not start server)
migrate:
	go run ./cmd/tools/migrate

# Run seeders (explicit, does not start server)
seed:
	go run ./cmd/tools/seed

# Run ERP Analyzer
# Usage: make analyze
# Usage: make analyze MODULES=finance,sales FROM=2025-01-01 TO=2025-12-31
analyze:
	@echo "Running Indosupplier ERP Analyzer..."
	@if [ -n "$(MODULES)" ] && [ -n "$(FROM)" ] && [ -n "$(TO)" ]; then \
		go run ./cmd/analyzer/main.go --module=$(MODULES) --from-date=$(FROM) --to-date=$(TO); \
	elif [ -n "$(MODULES)" ]; then \
		go run ./cmd/analyzer/main.go --module=$(MODULES); \
	elif [ -n "$(FROM)" ] && [ -n "$(TO)" ]; then \
		go run ./cmd/analyzer/main.go --from-date=$(FROM) --to-date=$(TO); \
	else \
		go run ./cmd/analyzer/main.go; \
	fi

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f server
	go clean

# Build Docker image
docker-build:
	docker build -t indosupplier-api .

# Run Docker container
docker-run:
	docker run -p 8080:8080 indosupplier-api

# Start with docker-compose
docker-up:
	docker-compose up --build

# Stop docker-compose
docker-down:
	docker-compose down

# Download dependencies
deps:
	go mod download
	go mod tidy

