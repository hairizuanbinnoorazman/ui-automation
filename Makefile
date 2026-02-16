.PHONY: build run test migrate-up migrate-down clean install-deps

BINARY_NAME=backend
CONFIG_FILE=config.yaml
MIGRATIONS_PATH=database/migrations

build:
	go build -o bin/$(BINARY_NAME) cmd/backend/*.go

run: build
	./bin/$(BINARY_NAME) serve -c $(CONFIG_FILE)

test:
	go test -v -race -cover ./...

migrate-up: build
	./bin/$(BINARY_NAME) migrate up -c $(CONFIG_FILE) -p $(MIGRATIONS_PATH)

migrate-down: build
	./bin/$(BINARY_NAME) migrate down -c $(CONFIG_FILE) -p $(MIGRATIONS_PATH)

clean:
	rm -rf bin/

install-deps:
	go mod download
	go mod tidy
