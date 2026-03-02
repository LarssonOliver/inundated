# This should probably be pinned in the future
OPENAPI_GENERATOR_TAG ?= latest-release

VERSION ?= $(shell git describe --tags --always --dirty)

BINARY_DIR := bin
BINARY := ${BINARY_DIR}/inundated

.PHONY: build dev build-frontend
build: ${BINARY}

dev:
	go run cmd/server/main.go

${BINARY}: build-frontend
	mkdir -p ${BINARY_DIR}
	go build -o $@ -tags embed cmd/server/main.go

build-frontend:
	cd frontend && npm install && npm run build

# Image builds
.PHONY: image-push
image-push:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=${VERSION} \
		--file build/Dockerfile \
		--tag larssonoliver/inundated:latest \
		--push .

# Run tests
.PHONY: test test-backend test-frontend
test: test-backend test-frontend

test-backend:
	@echo "==> Running backend tests..."
	go run gotest.tools/gotestsum@latest

test-frontend:
	@echo "==> Running frontend tests..."
	cd frontend && npm test

# Regenerate code from OpenAPI spec
.PHONY: generate generate-backend-api generate-frontend-api
generate: generate-backend-api generate-frontend-api

generate-backend-api:
	@echo "==> Running backend code generators..."
	go generate ./...

generate-frontend-api:
	@echo "==> Running frontend code generators..."
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_TAG} generate -i /local/openapi/inundated.yaml -g typescript-fetch --additional-properties=supportsES6=true -o /local/frontend/src/api/generated

