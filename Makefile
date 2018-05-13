NO_COLOR=\033[0m
OK_COLOR=\033[32;01m

all: deps lint test

deps:
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure

lint:
	@go vet ./...

test:
	@go test -v -cover ./...

checks:
	@docker-compose up -d
	@sleep 3
	@echo "Running checks tests against container deps" && \
		HEALTH_GO_PG_DSN="postgres://test:test@`docker-compose port postgres 5432`/test?sslmode=disable" \
		HEALTH_GO_MQ_DSN="amqp://guest:guest@`docker-compose port rabbit 5672`/" \
		HEALTH_GO_RD_DSN="redis://`docker-compose port redis 6379`/" \
		HEALTH_GO_MG_DSN="`docker-compose port mongo 27017`/" \
		go test -v -cover ./...

.PHONY: all deps test lint checks
