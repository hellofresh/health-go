all: lint test

lint:
	@go vet ./...

test:
	@go test -v -cover ./...

.PHONY: all test lint
