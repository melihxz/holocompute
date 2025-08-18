.PHONY: build test lint clean e2e bench

# Build the holo binary
build:
	go build -o bin/holo ./cmd/holo

# Run unit tests
test:
	go test -v ./...

# Run linting
lint:
	go vet ./...
	@echo "To run staticcheck, please install it first with: make install-tools"

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run end-to-end tests
e2e:
	# Placeholder for e2e tests
	echo "Running e2e tests"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install development tools
install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest