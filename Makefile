## Minimal helper Makefile (optional ergonomics, no heavy tooling)

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PORT    ?= 8080

.PHONY: build test cover run docker-build docker-run clean help

build: ## Compile all packages
	go build ./...

test: ## Run unit tests
	go test ./...

cover: ## Coverage (text)
	go test -cover ./...

run: ## Run example server locally
	go run ./examples/server -port $(PORT)

docker-build: ## Build container image with version metadata
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t go-sixtysix:$(VERSION) .

docker-run: ## Run container exposing PORT
	docker run --rm -e PORT=$(PORT) -p $(PORT):$(PORT) go-sixtysix:$(VERSION)

clean: ## Remove build cache (no artifacts kept here anyway)
	go clean -cache -testcache

help: ## Show targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | sed 's/:.*##/: /'
