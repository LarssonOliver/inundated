package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/stretchr/testify/assert"
)

func TestFrontendHandlerMethods(t *testing.T) {
	// Whether GET finds index.html depends on whether the frontend was built,
	// so only the method check is asserted.
	tests := map[string]bool{
		http.MethodGet:  true,
		http.MethodHead: true,
		http.MethodPost: false,
		"PROPFIND":      false,
	}
	for method, allowed := range tests {
		t.Run(method, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handlers.FrontendHandler().ServeHTTP(recorder, httptest.NewRequest(method, "/", nil))
			assert.Equal(t, allowed, recorder.Code != http.StatusMethodNotAllowed)
		})
	}
}
