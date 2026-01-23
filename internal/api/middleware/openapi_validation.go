package middleware

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/larssonoliver/inundated/openapi"

	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

var spec *openapi3.T = nil

func init() {
	s, err := openapi3.NewLoader().LoadFromData(openapi.OpenAPISpec)
	if err != nil {
		panic(err)
	}

	// NOTE that we need to make sure that the `Servers` aren't set, otherwise the
	// OpenAPI validation middleware will validate that the `Host` header
	// (of incoming requests) are targeting known `Servers` in the OpenAPI spec
	// See also: Options#SilenceServersWarning
	s.Servers = nil

	spec = s
}

func OpenApiRequestValidator() func(http.Handler) http.Handler {
	return nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		Options: openapi3filter.Options{},
	})
}
