.PHONY: build dev clean install-deps serve

# Build the Elm application
build:
	elm make src/App.elm --output=elm.js

# Build with optimizations for production
build-prod:
	elm make src/App.elm --output=elm.js --optimize

# Development mode with live reload (requires elm-live)
dev:
	elm-live src/App.elm --open -- --output=elm.js

# Clean build artifacts
clean:
	rm -f elm.js elm.min.js
	rm -rf elm-stuff

# Install development dependencies
install-deps:
	npm install -g elm elm-live

# Serve the application using Python's HTTP server
serve: build
	python3 -m http.server 8000

# Install Elm packages
elm-install:
	elm make src/App.elm

# Format Elm code
format:
	elm-format src/ --yes
