SHELL := /bin/bash

all: server test

server:
	cd cmd/server && go build -o ncdmv_server .
	mv cmd/server/ncdmv_server .

test:
	go test -v -short ./...

clean:
	rm ncdmv_server

.PHONY: all clean test

