SHELL := /bin/bash

all: server test

server: db-gen
	cd cmd/server && go build -o ncdmv_server .

test:
	go test -v -short ./...

clean:
	rm -f cmd/server/ncdmv_server

# Re-generate SQL models from the schema.
db-gen:
	docker run --rm \
		-v $(CURDIR):/src -w /src \
		kjconroy/sqlc generate

# Run all up migrations. Requires the NCDMV_DB_PATH env var to be set.
# Example:
# NCDMV_DB_PATH="./ncdmv.db" make db-migrate
db-migrate:
	go run ./cmd/migrate

db-clean:
	rm -f database/ncdmv.db

.PHONY: all clean db-clean db-gen db-migrate test
