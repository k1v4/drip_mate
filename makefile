GO := go
GOLANGCI_LINT := golangci-lint

.PHONY: all format lint

all: format lint

format:
	$(GO) fmt ./...

lint:
	$(GOLANGCI_LINT) run --timeout=5m ./... 
