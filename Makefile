OK_COLOR=\033[32;01m
NO_COLOR=\033[0m

test:
	@echo "$(OK_COLOR)==> Running tests against container deps$(NO_COLOR)"
	@docker-compose up -d
	@sleep 3 && \
		HEALTH_GO_PG_PQ_DSN="postgres://test:test@`docker-compose port pg-pq 5432`/test?sslmode=disable" \
		HEALTH_GO_PG_PGX4_DSN="postgres://test:test@`docker-compose port pg-pgx4 5432`/test?sslmode=disable" \
		HEALTH_GO_MQ_DSN="amqp://guest:guest@`docker-compose port rabbit 5672`/" \
		HEALTH_GO_MQ_URL="http://guest:guest@`docker-compose port rabbit 15672`/" \
		HEALTH_GO_RD_DSN="redis://`docker-compose port redis 6379`/" \
		HEALTH_GO_MG_DSN="mongodb://`docker-compose port mongo 27017`/" \
		HEALTH_GO_MS_DSN="test:test@tcp(`docker-compose port mysql 3306`)/test?charset=utf8" \
		HEALTH_GO_HTTP_URL="http://`docker-compose port http 8080`/status" \
		HEALTH_GO_MD_DSN="memcached://localhost:${{ job.services.memcached.ports[11211] }}/" \
		HEALTH_GO_INFLUXDB_URL="http://`docker-compose port influxdb 8086`" \
		go test -cover ./... -coverprofile=coverage.txt -covermode=atomic

lint:
	@echo "$(OK_COLOR)==> Linting with golangci-lint$(NO_COLOR)"
	@docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.33.0 golangci-lint run -v

.PHONY: test lint
