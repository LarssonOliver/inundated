
# Regenerate code from OpenAPI spec
.PHONY: generate
generate:
	@echo "==> Running code generators..."
	go generate ./...
	openapi-generator generate -i ./openapi/inundated.yaml -g typescript-fetch --additional-properties=supportsES6=true -o ./frontend/src/api/

