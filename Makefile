# This should probably be pinned in the future
OPENAPI_GENERATOR_TAG ?= latest-release

BINARY_DIR := bin
BINARY := ${BINARY_DIR}/inundated

.PHONY: build
build: ${BINARY}

${BINARY}: build-frontend
	mkdir -p ${BINARY_DIR}
	go build -o $@ cmd/server/main.go

build-frontend:
	cd frontend && npm install && npm run build

# Run tests
.PHONY: test test-backend test-frontend
test: test-backend test-frontend

test-backend:
	@echo "==> Running backend tests..."
	go test -v ./...

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
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_TAG} generate -i /local/openapi/inundated.yaml -g typescript-fetch --additional-properties=supportsES6=true -o /local/frontend/src/api/

