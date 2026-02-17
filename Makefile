.PHONY: build run test migrate-up migrate-down clean install-deps docker-dev docker-build-elm docker-check-elm docker-rebuild-elm

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

# Check if elm.js exists, build if missing
docker-check-elm:
	@if [ ! -f frontend/elm.js ]; then \
		echo "elm.js not found. Building..."; \
		cd frontend && elm make src/App.elm --output=elm.js; \
	else \
		echo "elm.js found ($$(ls -lh frontend/elm.js | awk '{print $$5}'))"; \
	fi

# Build elm.js for development
docker-build-elm:
	cd frontend && elm make src/App.elm --output=elm.js

# Start Docker Compose with dev setup (checks for elm.js first)
docker-dev: docker-check-elm
	docker compose up

# Rebuild elm.js and restart frontend container
docker-rebuild-elm: docker-build-elm
	docker compose restart frontend
