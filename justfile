# NOTE: RUN "just -l" TO QUICKLY VIEW ALL COMMANDS

alias genmg := gen-migration-files
[group('gen')]
gen-migration-files:
    @migrate create -ext sql -dir migrations -seq create_entries_table

# generate a random password 32 characters long for database
[group('gen')]
genpass:
    @openssl rand -hex 32

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
