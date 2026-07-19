# NOTE: RUN "just -l" TO QUICKLY VIEW ALL COMMANDS

set dotenv-load

db_url := "postgres://" + env_var("DB_USER") + ":" + env_var("DB_PASSWORD") + "@localhost:" + env_var("DB_PORT") + "/" + env_var("DB_NAME") + "?sslmode=disable"

alias genmg := gen-migration-files
[group('migrations')]
gen-migration-files:
    @migrate create -ext sql -dir migrations -seq create_entries_table

alias mgup := migrate-up
[group('migrations')]
migrate-up:
    @migrate -path migrations -database "{{db_url}}" up

alias mgdown := migrate-down
# rollback to the last migration
[group('migrations')]
migrate-down:
    @migrate -path migrations -database "{{db_url}}" down 1

alias mgver := migrate-version
[group('migrations')]
migrate-version:
    @migrate -path migrations -database "{{db_url}}" version

# generate a random password 32 characters long for database
[group('gen')]
genpass:
    @openssl rand -hex 32

# gen code from sql
[group('gen')]
sqlc:
    @sqlc generate

[group('docker')]
up:
    @docker-compose up -d

[group('docker')]
down:
    @docker-compose down

# quick view db tables instead of opening db editor
[group('docker')]
checkdb:
    @docker exec -it log-pipeline-db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\d entries"'

# quick view db data instead of opening db editor
[group('docker')]
showdb:
    @docker exec -it log-pipeline-db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT id, type, service, message, created_at FROM entries ORDER BY id DESC LIMIT 5;"'

alias rcli := redis-cli
[group('redis')]
redis-cli:
    @docker exec -it log-pipeline-redis redis-cli

# redis xrange
[group('redis')]
xrange:
    @docker exec -it log-pipeline-redis redis-cli XRANGE ingest_stream - +

# redis xlen
[group('redis')]
xlen:
    @docker exec -it log-pipeline-redis redis-cli XLEN ingest_streamv

# run linter
[group('dev')]
lint:
    @golangci-lint run ./...

# run all tests
[group('dev')]
test:
    @go test ./... -v

# run both ingest and consumer concurrently
[group('dev')]
dev:
    @air -c .air.ingest.toml & air -c .air.consumer.toml & wait
