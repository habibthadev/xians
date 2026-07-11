.PHONY: dev-api dev-web build-api build-web test-api test-web lint-api lint fmt clean

# Go variables
GO ?= go
GOFLAGS ?=
BIN_DIR ?= $(CURDIR)/bin

dev-api:
	mkdir -p $(BIN_DIR)
	cd apps/api && $(GO) build $(GOFLAGS) -o $(BIN_DIR)/xians-api ./cmd/server
	$(BIN_DIR)/xians-api

dev-web:
	pnpm --filter @xians/web dev

build-api:
	cd apps/api && $(GO) build $(GOFLAGS) -o $(BIN_DIR)/xians-api ./cmd/server

build-reaper:
	cd apps/api && $(GO) build $(GOFLAGS) -o $(BIN_DIR)/xians-reaper ./cmd/reaper

build-web:
	pnpm --filter @xians/web build

build: build-api build-web

test-api:
	cd apps/api && $(GO) test -v -race -count=1 ./...

test-web:
	pnpm --filter @xians/web check

test: test-api test-web

lint-api:
	cd apps/api && golangci-lint run ./...

lint: lint-api
	pnpm -r lint

fmt:
	cd apps/api && $(GO) fmt ./...
	cd apps/api && gofumpt -w ./...

clean:
	rm -rf $(BIN_DIR)
	rm -rf apps/web/.svelte-kit apps/web/build
