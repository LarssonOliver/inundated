package middleware

import (
	"net/http"

	"github.com/larssonoliver/inundated/internal/service"
)

func OIDCAuth(userService service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			next.ServeHTTP(w, r)
		})
	}
}

