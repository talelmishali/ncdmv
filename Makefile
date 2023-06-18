SHELL := /bin/bash

all: db-gen server test

server:
	go build -o ncdmv ./cmd/server

test:
	go test -v -short ./...

install:
	go build -v -o /usr/local/bin/ncdmv ./cmd/server

clean:
	rm -f ./ncdmv

docker:
	docker build . -t ncdmv

# Re-generate SQL models from the schema.
db-gen:
	docker run --rm \
		-v $(CURDIR):/src -w /src \
		kjconroy/sqlc generate

# Run all up migrations. Requires the NCDMV_DB_PATH env var to be set.
# Example:
# NCDMV_DB_PATH="./ncdmv.db" make db-migrate
db-migrate:
	go run ./cmd/migrate && mv database/ncdmv.db .

db-clean:
	rm -f ./ncdmv.db
	rm -f ./database/ncdmv.db

.PHONY: all clean db-clean db-gen db-migrate docker install server test
