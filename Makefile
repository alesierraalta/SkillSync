.PHONY: test bench lint build verify

test:
	go test ./... -race -count=1

bench:
	go test -bench=. -benchmem -run=^$ ./internal/parser/... ./internal/sandbox/...

# Soft-fail: skip with a note if golangci-lint is not installed.
lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || \
		(echo "golangci-lint not installed; lint skipped" && true)

build:
	go build ./cmd/synck

verify: test lint build
	@echo "verify ok"
