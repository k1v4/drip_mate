GO := go
GOLANGCI_LINT := golangci-lint
MOCKER := mockery
GOIMPORTS := goimports

.PHONY: all generate-mocks format lint

all: generate-mocks format lint

generate-mocks:
	$(MOCKER) --config=mockery.yml

format:
	$(GO) fmt ./...
	$(GOIMPORTS) -w .

lint:
	$(GOLANGCI_LINT) run --timeout=5m ./...
