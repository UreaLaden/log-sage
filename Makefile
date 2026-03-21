.PHONY: run-lint fmt lint test-install pack-choco

CHOCO_REPO ?= /tmp/chocolatey-logsage
CHOCO_OUT ?= $(CHOCO_REPO)/dist
CHOCO ?= choco.exe

fmt:
	gofmt -w .

lint:
	golangci-lint run --timeout=5m ./...

run-lint: fmt lint

test-install:
	sh scripts/test-install.sh

pack-choco:
	@test -f "$(CHOCO_REPO)/logsage.nuspec" || { echo "missing $(CHOCO_REPO)/logsage.nuspec"; exit 1; }
	@mkdir -p "$(CHOCO_OUT)"
	@"$(CHOCO)" pack "$(CHOCO_REPO)/logsage.nuspec" --outputdirectory "$(CHOCO_OUT)"
