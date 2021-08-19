module github.com/hellofresh/health-go/v4

go 1.15

require (
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/go-redis/redis/v8 v8.10.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jackc/pgx/v4 v4.11.0
	github.com/lib/pq v1.10.0
	github.com/rabbitmq/amqp091-go v0.0.0-20210812094702-b2a427eb7d17
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.7.1
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/internal/metric v0.22.0 // indirect
	go.opentelemetry.io/otel/trace v1.0.0-RC2
	google.golang.org/grpc v1.36.0
)
