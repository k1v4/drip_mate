GO := go
GOLANGCI_LINT := golangci-lint
MOCKER := mockery
GOIMPORTS := goimports
COVERPKGS := $(shell go list ./... | grep -v mocks | grep -v repository | grep -v config | grep -v cmd | paste -sd ',' -)

COVERAGE_MIN := 70

.PHONY: all format lint test

all: format lint test

generate-mocks:
	$(MOCKER) --config=mockery.yml

format:
	$(GO) fmt ./...
	$(GO) fix ./...
	$(GOIMPORTS) -w .

lint:
	$(GOLANGCI_LINT) run --timeout=5m ./... --fix

test:
	@echo "Running tests with coverage"
	$(GO) test ./... -coverpkg=$(COVERPKGS) -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@coverage=$$($(GO) tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Total coverage: $$coverage%"; \
	awk "BEGIN { exit !($$coverage >= $(COVERAGE_MIN)) }" || \
	( echo "Coverage $$coverage% is below $(COVERAGE_MIN)%"; exit 1 )
