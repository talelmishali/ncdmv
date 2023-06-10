SHELL := /bin/bash

all: server test

server:
	cd cmd/server && go build -o ncdmv_server .

test:
	go test -v -short ./...

clean:
	rm -f cmd/server/ncdmv_server

# Run all up migrations. Requires the NCDMV_DB_PATH env var to be set.
# Example:
# NCDMV_DB_PATH="./ncdmv.db" make db-migrate
db-migrate:
	go run ./cmd/migrate

.PHONY: all clean db-migrate test
