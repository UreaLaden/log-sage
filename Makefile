.PHONY: run-lint fmt lint

fmt:
	gofmt -w .

lint:
	golangci-lint run --timeout=5m ./...

run-lint: fmt lint
