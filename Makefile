# Inventory API — common developer commands.
# Usage: `make run`, `make build`, `make tidy`, etc.

# Default target when you just type `make`.
.DEFAULT_GOAL := run

# run: start the API server (development).
run:
	go run ./cmd/api

# build: compile a single production binary into ./bin/api
build:
	go build -o bin/api ./cmd/api

# tidy: clean up go.mod / go.sum (add missing, remove unused deps).
tidy:
	go mod tidy

# fmt: auto-format all Go code (like Laravel Pint).
fmt:
	go fmt ./...

# vet: static analysis — catches suspicious code before it ships.
vet:
	go vet ./...

# swagger: regenerate the OpenAPI docs from handler annotations.
# Requires: go install github.com/swaggo/swag/cmd/swag@latest
swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# .PHONY tells make these are command names, not files to produce.
.PHONY: run build tidy fmt vet swagger
