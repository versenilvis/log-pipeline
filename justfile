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

[group('gen')]
sqlc:
    @sqlc generate

[group('docker')]
up:
    @docker-compose up -d

[group('docker')]
down:
    @docker-compose down

# quick view db instead of opening db editor
[group('docker')]
checkdb:
    @docker exec -it log-pipeline-db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\d entries"'
