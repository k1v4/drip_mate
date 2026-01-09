GO := go
GOLANGCI_LINT := golangci-lint
MOCKER := mockery

.PHONY: all generate-mocks format lint

all: generate-mocks format lint

generate-mocks:
	./generate-mockery-yaml.sh
	$(MOCKER) --config=mockery.yml

format:
	$(GO) fmt ./...

lint:
	$(GOLANGCI_LINT) run --timeout=5m ./...
