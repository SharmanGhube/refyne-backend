.PHONY: wire
wire:
	@echo "Running wire..."
	cd cmd && wire

.PHONY: build
build: wire
	@echo "Building the application..."
	go build -o bin/app ./cmd

.PHONY: ensure-redis
ensure-redis:
	@echo "Ensuring Redis is running..."
	-docker compose up -d redis

.PHONY: run
run: build ensure-redis
	@echo "Running the application..."
	./bin/app

.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	del /f cmd\wire_gen.go 2>nul || true
	rmdir /s /q bin 2>nul || true

.PHONY: dev
dev: wire
	air -c .air.toml

.PHONY: test
test:
	@echo "Running tests..."
	go test ./... -v

.PHONY: test-short
test-short:
	@echo "Running tests (short mode)..."
	go test ./... -short

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-coverage-func
test-coverage-func:
	@echo "Running tests with coverage summary..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...