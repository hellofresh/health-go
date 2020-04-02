OK_COLOR=\033[32;01m
NO_COLOR=\033[0m

all: deps test

deps:
	@echo "$(OK_COLOR)==> Installing dependencies $(NO_COLOR)"
	@go install -i golang.org/x/lint/golint
	@go mod vendor

code-quality: lint format vet

test: code-quality
	@echo "$(OK_COLOR)==> Running tests against container deps $(NO_COLOR)"
	@docker-compose up -d
	@sleep 3 && \
		HEALTH_GO_PG_DSN="postgres://test:test@`docker-compose port postgres 5432`/test?sslmode=disable" \
		HEALTH_GO_MQ_DSN="amqp://guest:guest@`docker-compose port rabbit 5672`/" \
		HEALTH_GO_MQ_URL="http://guest:guest@`docker-compose port rabbit 15672`/" \
		HEALTH_GO_RD_DSN="redis://`docker-compose port redis 6379`/" \
		HEALTH_GO_MG_DSN="mongodb://`docker-compose port mongo 27017`/" \
		HEALTH_GO_MS_DSN="test:test@tcp(`docker-compose port mysql 3306`)/test?charset=utf8" \
		HEALTH_GO_HTTP_URL="http://`docker-compose port http 8080`/status" \
		go test -cover ./... -coverprofile=coverage.txt -covermode=atomic

lint:
	@echo "$(OK_COLOR)==> Checking code style with 'golint' tool$(NO_COLOR)"
	@go list ./... | xargs -n 1 golint -set_exit_status

format:
	@echo "$(OK_COLOR)==> Checking code formating with 'gofmt' tool$(NO_COLOR)"
	@gofmt -l -s cmd pkg | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

vet:
	@echo "$(OK_COLOR)==> Checking code correctness with 'go vet' tool$(NO_COLOR)"
	@go vet ./...

.PHONY: all deps test code-quality lint format vet
