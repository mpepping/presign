VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w \
	-X github.com/mpepping/presign/internal/cmd.version=$(VERSION) \
	-X github.com/mpepping/presign/internal/cmd.commit=$(COMMIT) \
	-X github.com/mpepping/presign/internal/cmd.date=$(DATE)

.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	go build -buildvcs=false -ldflags "$(LDFLAGS)" -o bin/presign ./cmd/presign

install: ## Install the binary to GOPATH/bin
	go install -ldflags "$(LDFLAGS)" ./cmd/presign

test: ## Run tests
	go test ./...

lint: ## Run linters
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/

