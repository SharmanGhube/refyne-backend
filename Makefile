.PHONY: wire
wire:
	@echo "Running wire..."
	cd cmd && wire

.PHONY: build
build: wire
	@echo "Building the application..."
	go build -o bin/app ./cmd

.PHONY: run
run: build
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