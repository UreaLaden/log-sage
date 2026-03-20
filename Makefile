.PHONY: run-lint fmt lint test-install

fmt:
	gofmt -w .

lint:
	golangci-lint run --timeout=5m ./...

run-lint: fmt lint

test-install:
	sh scripts/test-install.sh
