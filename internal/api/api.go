//go:generate go tool oapi-codegen --config oapi-codegen/model.cfg.yaml ../../openapi/inundated.yaml
//go:generate go tool oapi-codegen --config oapi-codegen/server.cfg.yaml ../../openapi/inundated.yaml

package api

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}
